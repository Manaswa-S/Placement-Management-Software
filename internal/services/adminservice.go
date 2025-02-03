package services

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mod/internal/apicalls"
	sqlc "go.mod/internal/sqlc/generate"
	"go.mod/internal/utils"
)

type AdminService struct {
	queries *sqlc.Queries
	GAPIService *apicalls.Caller
}
func NewAdminService(queriespool *sqlc.Queries, gapiService *apicalls.Caller) *AdminService {
	return &AdminService{
		queries: queriespool,
		GAPIService: gapiService,
	}
}

func (a *AdminService) AdminFunc(ctx *gin.Context) () {
}

func (a *AdminService) StudentInfo(ctx *gin.Context, userid string) (*sqlc.StudentInfoRow, error) {

	userID, err := strconv.ParseInt(userid, 10, 64)
	if err != nil {
		return nil, err
	}

	sData, err := a.queries.StudentInfo(ctx, userID)
	if err != nil {
		return nil, err
	}

	ppByte, err := os.ReadFile(sData.Profilepic.String)
	if err != nil {
		fmt.Println(err)
	}

	sData.Profilepic.String = base64.StdEncoding.EncodeToString(ppByte)

	return &sData, nil
}


func (a *AdminService) ManageStudents(ctx *gin.Context, tab string) (any, error) {


	switch tab {
		case "overview":
			allData, err := a.queries.StudentsOverview(ctx)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			return allData, nil

		case "verify":
			toVerify, err := a.queries.ListToVerifyStudent(ctx)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			return toVerify, nil
	}




	return nil, nil
}
func (a *AdminService) VerifyStudent(ctx *gin.Context, userid string) (error) {
	userID, err := strconv.ParseInt(userid, 10, 64)
	if err != nil {
		return err
	}

	err = a.queries.VerifyStudent(ctx, userID)
	if err != nil {	
		return err
	}

	// TODO: notify student of verification

	return nil
}

func (a *AdminService) GenerateTestResult(ctx *gin.Context, testid string) (error) {
	// parse test id 
	testID, err := strconv.ParseInt(testid, 10, 64)
	if err != nil {
		return err
	}



	resultPath, err := utils.GenerateTestResultDraft(a.queries, a.GAPIService, testID)
	if err != nil {
		return err
	}
	fmt.Println(resultPath)

	return nil
}
