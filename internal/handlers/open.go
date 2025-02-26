package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.mod/internal/config"
	errs "go.mod/internal/const"
	"go.mod/internal/dto"
	"go.mod/internal/services"
	"go.mod/internal/utils/ctxutils"
)

type OpenHandler struct {
	OpenService *services.OpenService
}

func NewOpenHandler(openService *services.OpenService) *OpenHandler {
	return &OpenHandler{
		OpenService: openService,
	}
}

func (h *OpenHandler) RegisterRoute(openRoute *gin.RouterGroup) {
	openRoute.GET("/discussions", h.Discussions)
	openRoute.GET("/discussionsdata", h.DiscussionsData)
	openRoute.POST("/newdiscussion", h.NewDiscussion)
	openRoute.POST("/newreply", h.NewReply)
	openRoute.GET("/replies", h.GetReplies)
}


// extractUserID extracts the user ID and other required parameters from the context with explicit type assertion.
// any returned error is directly included in the response as returned
func (h *OpenHandler) extractUserID(ctx *gin.Context) (int64, *errs.Error) {

	userid, exists := ctx.Get("ID")
	if !exists {
		return 0, &errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing user ID in request.",
			ToRespondWith: true,
		}
	}

	userID, ok := userid.(int64)
	if !ok {
		return 0, &errs.Error{
			Type: errs.InvalidFormat,
			Message: "User ID of improper format.",
			ToRespondWith: true,
		}
	}

	return userID, nil 
}
// checkFile checks file validity/existence for the given filePath.
// It also does ctx.Set(error) and returns a structured *errs.Error object too for any errors
func (h *OpenHandler) checkFile(ctx *gin.Context, filePath string) *errs.Error {

	if filePath == "" {
		ctx.Set("error", "filePath is empty string.")
		return &errs.Error{
			Type: errs.NotFound,
			Message: "filePath is empty.",
		}
	}

	_, err := os.Stat(filePath)
	if err != nil {
		ctx.Set("error", "cannot get file data for path : " + filePath)
		return &errs.Error{
			Type: errs.NotFound,
			Message: "Failed to get metadata/ access file : " + filePath,
		}
	}

	return nil
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

func (h *OpenHandler) Discussions(ctx *gin.Context) {

	filePath := config.OpenPaths.DiscussionsPagePath
	errf := h.checkFile(ctx, filePath)
	if errf != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.File(filePath)
}

func (h *OpenHandler) DiscussionsData(ctx *gin.Context) {
	page := ctx.Query("page")
	if page == "" {
		return
	}

	userID, errf := ctxutils.ExtractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	data, limit, offset, err := h.OpenService.DiscussionsData(ctx, userID, page)
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"Data": data,
		"Limit": limit,
		"Offset": offset,
		"RefreshRate": config.DiscussionPageAutoRefreshRate,
	})
}

func (h *OpenHandler) NewDiscussion(ctx *gin.Context) {
	
	data := new(dto.NewDiscussion)

	err := ctx.Bind(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	userid, exists := ctx.Get("ID")
	if !exists {
		return
	}

	errf := h.OpenService.NewDiscussion(ctx, userid.(int64), data)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"Status": "Posted new discussion successfully.",
	})
}

func (h *OpenHandler) NewReply(ctx *gin.Context) {

	data := new(dto.NewReply)
	err := ctx.Bind(data)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.IncompleteForm,
			Message: "New reply form is incomplete or invalid.",
			ToRespondWith: true,
		})
		return
	}

	userID, errf := ctxutils.ExtractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	errf = h.OpenService.NewReply(ctx, userID, data)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"Status": "Replied successfully.",
	})
}

func (h *OpenHandler) GetReplies(ctx *gin.Context) {

	postid := ctx.Query("postid")
	if postid == "" {
		return
	}

	replies, errf := h.OpenService.GetReplies(ctx, postid)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, replies)
}

















func (h *OpenHandler) EditPost(ctx *gin.Context) {
	data := new(dto.EditDiscussion)
	err := ctx.Bind(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	errf = h.OpenService.EditDiscussion(ctx, userID, data)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"Status": "Edited discussion successfully.",
	})
}
