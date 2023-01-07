package main

import (
	"context"
	"corona-api/src/modules/patient"
	"fmt"
	"sync"
)

const GeneratePatientDetailsResponseParallelNumber = 1

func main() {
	// 準備
	patientDetails := []patient.Detail{
		{Date: 202301, Area: "北海道", Value: 1000, Country: "日本"},
	}

	result, err := MainGeneratePatientDetailsResponse(patientDetails)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}

func MainGeneratePatientDetailsResponse(patientDetails []patient.Detail) ([]byte, error) {
	// メイン
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

	wg.Wait()
	if ctx.Err() != nil {
		return nil, <-errorCh
	}

	area := <-areaCh
	fmt.Println(area)

	return []byte{}, nil
}

func createArea(patientDetails []patient.Detail) (string, error) {
	return "", fmt.Errorf("errorです！")
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
