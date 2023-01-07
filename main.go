package main

import (
	docs "corona-api/docs"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
)

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @host      localhost:8081
func main() {
	r := gin.Default()
	docs.SwaggerInfo.BasePath = ""
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	r.Run(":8081")
}

func test(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

// @BasePath /api/v1

// PingExample godoc
// @Summary ping example
// @Schemes
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /example/helloworld [get]
func Helloworld(g *gin.Context) {
	g.JSON(http.StatusOK, "helloworld")
}

//func MainGeneratePatientDetailsResponse(patientDetails []patient.Detail) ([]byte, error) {
//	// メイン
//	var wg sync.WaitGroup
//	wg.Add(GeneratePatientDetailsResponseParallelNumber)
//
//	// キャンセル機能
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	// エラーチャンネル
//	errorCh := make(chan error, GeneratePatientDetailsResponseParallelNumber)
//	defer close(errorCh)
//
//	// エリア取得
//	areaCh := make(chan string, 1)
//	defer close(areaCh)
//	go func() {
//		defer wg.Done()
//		select {
//		case <-ctx.Done():
//			return
//		default:
//			area, err := createArea(patientDetails)
//			if err != nil {
//				errorCh <- err
//				cancel()
//				return
//			}
//			areaCh <- area
//		}
//	}()
//
//	wg.Wait()
//	if ctx.Err() != nil {
//		return nil, <-errorCh
//	}
//
//	area := <-areaCh
//	fmt.Println(area)
//
//	return []byte{}, nil
//}
//
//func createArea(patientDetails []patient.Detail) (string, error) {
//	return "", fmt.Errorf("errorです！")
//	var area string
//	for _, pd := range patientDetails {
//		if pd.Area == "" {
//			return "", fmt.Errorf("area is empty")
//		}
//		area = pd.Area
//		return area, nil
//	}
//	return area, nil
//}
