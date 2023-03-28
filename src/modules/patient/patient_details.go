package patient

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
)

const (
	GeneratePatientDetailsResponseParallelNumber = 3
)

type Detail struct {
	Date    uint32
	Area    string
	Value   uint32
	Country string
}

func GetPatientDetailsByPeriodAndArea(db *sql.DB, area string, startDate uint32, endDate uint32) ([]Detail, error) {
	rows, err := db.Query("SELECT  date, area, value, country FROM patient_details WHERE area = ? AND date BETWEEN ? AND ?", area, startDate, endDate)
	if err != nil {
		return []Detail{}, fmt.Errorf("db.Query() error: %v", err)
	}
	defer rows.Close()

	var patientDetails []Detail
	for rows.Next() {
		var patientDetail Detail
		err := rows.Scan(&patientDetail.Date, &patientDetail.Area, &patientDetail.Value, &patientDetail.Country)
		if err != nil {
			return []Detail{}, fmt.Errorf("rows.Scan() error: %v", err)
		}
		patientDetails = append(patientDetails, patientDetail)
	}
	err = rows.Err()
	if err != nil {
		return []Detail{}, fmt.Errorf("rows.Err() error: %v", err)
	}
	return patientDetails, nil
}

func GeneratePatientDetailsResponse(patientDetails []Detail) ([]byte, error) {
	body := map[string]interface{}{}

	// 日付データを作成
	body = createDate(patientDetails, body)

	// 並行処理でデータを取得
	var wg sync.WaitGroup
	wg.Add(GeneratePatientDetailsResponseParallelNumber)

	// キャンセル機能
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// エラーチャンネル
	errorCh := make(chan error, GeneratePatientDetailsResponseParallelNumber)
	defer close(errorCh)

	// エリア取得
	areaCh := make(chan string, 1)
	defer close(areaCh)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			return
		default:
			area, err := createArea(patientDetails)
			if err != nil {
				errorCh <- err
				cancel()
				return
			}
			areaCh <- area
		}
	}()

	// 合計データを作成
	sumCh := make(chan uint32, 1)
	defer close(sumCh)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			return
		default:
			sum := createSum(patientDetails)
			sumCh <- sum
		}
	}()

	// 平均データを作成
	averageCh := make(chan float64, 1)
	defer close(averageCh)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			return
		default:
			average := createAverage(patientDetails)
			averageCh <- average
		}
	}()

	wg.Wait()
	if ctx.Err() != nil {
		return nil, <-errorCh
	}
	body["area"] = <-areaCh
	body["sum"] = <-sumCh
	body["average"] = <-averageCh

	// JSONにして返却
	bytes, err := json.Marshal(body)
	if err != nil {
		return []byte{}, fmt.Errorf("JSON marshal error: body: %v, %v ", body, err)
	}
	return bytes, nil
}

func createArea(patientDetails []Detail) (string, error) {
	var area string
	for _, pd := range patientDetails {
		if pd.Area == "" {
			return "", fmt.Errorf("area is empty")
		}
		area = pd.Area
		return area, nil
	}
	return area, nil
}

func createDate(patientDetails []Detail, body map[string]interface{}) map[string]interface{} {
	for _, pd := range patientDetails {
		dateString := strconv.Itoa(int(pd.Date))
		body[dateString] = pd.Value
	}
	return body
}

func createSum(patientDetails []Detail) uint32 {
	var sum uint32
	for _, pd := range patientDetails {
		sum += pd.Value
	}
	return sum
}

func createAverage(patientDetails []Detail) float64 {
	patientDetailsLength := len(patientDetails)
	var sum uint32
	for _, pd := range patientDetails {
		sum += pd.Value
	}
	return float64(sum) / float64(patientDetailsLength)
}

func InsertPatientDetails(db *sql.DB, patientDetails []Detail) error {
	// DBトランザクション開始
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		return err
	}

	// DBをTRUNCATE
	patientDetailsTableName := "patient_details"
	_, err = tx.Exec("TRUNCATE " + patientDetailsTableName)
	if err != nil {
		return err
	}

	// DBに保存
	stmt, err := tx.Prepare("INSERT INTO  patient_details (date, area, value, country) VALUES (?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, pd := range patientDetails {
		_, err := stmt.Exec(pd.Date, pd.Area, pd.Value, pd.Country)
		if err != nil {
			return err
		}
	}

	// SQLが正常に実行できたらコミット
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
