package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artnikel/blogapi/internal/handler/mocks"
	"github.com/artnikel/blogapi/internal/model"
	"github.com/artnikel/blogapi/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gopkg.in/go-playground/validator.v9"
)

func Test_Create(t *testing.T) {
	mockService := new(mocks.MockBlogService)
	validate := validator.New()
	h := NewHandler(mockService, nil, validate)

	blogInput := model.Blog{
		BlogID:  uuid.New(),
		Title:   "testtitle",
		Content: "testcontent",
	}
	bodyBytes, err := json.Marshal(blogInput)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/blog", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userID := uuid.New()
	c.Set("id", userID)

	mockService.On("Create", mock.Anything, mock.MatchedBy(func(b *model.Blog) bool {
		return b.Title == blogInput.Title && b.Content == blogInput.Content && b.UserID == userID && b.BlogID != uuid.Nil
	})).Return(nil)

	err = h.Create(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, rec.Code)

	var respBlog model.Blog
	err = json.Unmarshal(rec.Body.Bytes(), &respBlog)
	require.NoError(t, err)
	require.Equal(t, blogInput.Title, respBlog.Title)
	require.Equal(t, blogInput.Content, respBlog.Content)
	require.Equal(t, userID, respBlog.UserID)

	mockService.AssertExpectations(t)
}

func Test_Get(t *testing.T) {
	mockService := new(mocks.MockBlogService)
	validate := validator.New()
	h := NewHandler(mockService, nil, validate)

	id := uuid.New()
	expectedBlog := &model.Blog{
		BlogID:  id,
		Title:   "testtitle",
		Content: "testcontent",
	}

	mockService.On("Get", mock.Anything, id).Return(expectedBlog, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/blog/"+id.String(), http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())

	err := h.Get(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	var respBlog model.Blog
	err = json.Unmarshal(rec.Body.Bytes(), &respBlog)
	require.NoError(t, err)
	require.Equal(t, expectedBlog, &respBlog)

	mockService.AssertExpectations(t)
}

func Test_Delete_AsAdmin(t *testing.T) {
	mockService := new(mocks.MockBlogService)
	validate := validator.New()
	h := NewHandler(mockService, nil, validate)

	id := uuid.New()

	mockService.On("Delete", mock.Anything, id).Return(nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/blog/"+id.String(), http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set("isAdmin", true)
	err := h.Delete(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Deleted: "+id.String())

	mockService.AssertExpectations(t)
}

func Test_Delete_AsUserOwnBlog(t *testing.T) {
	mockService := new(mocks.MockBlogService)
	validate := validator.New()
	h := NewHandler(mockService, nil, validate)

	userID := uuid.New()
	blogID := uuid.New()

	blogs := []*model.Blog{
		{
			BlogID: blogID,
		},
	}

	mockService.On("GetByUserID", mock.Anything, userID).Return(blogs, nil)
	mockService.On("Delete", mock.Anything, blogID).Return(nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/blog/"+blogID.String(), http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(blogID.String())
	c.Set("id", userID)

	err := h.Delete(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Deleted: "+blogID.String())

	mockService.AssertExpectations(t)
}

func Test_Delete_NotOwner(t *testing.T) {
	mockService := new(mocks.MockBlogService)
	validate := validator.New()
	h := NewHandler(mockService, nil, validate)

	userID := uuid.New()
	blogID := uuid.New()

	blogs := []*model.Blog{}

	mockService.On("GetByUserID", mock.Anything, userID).Return(blogs, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/blog/"+blogID.String(), http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(blogID.String())
	c.Set("id", userID)

	err := h.Delete(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Contains(t, rec.Body.String(), "Cannot delete blog with id: "+blogID.String())

	mockService.AssertExpectations(t)
}

func Test_DeleteByUserID_SameUser(t *testing.T) {
	mockService := new(mocks.MockBlogService)
	validate := validator.New()
	h := NewHandler(mockService, nil, validate)

	userID := uuid.New()

	mockService.On("DeleteByUserID", mock.Anything, userID).Return(nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/blogs/user/"+userID.String(), http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(userID.String())
	c.Set("id", userID)

	err := h.DeleteByUserID(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Deleted from user id: "+userID.String())

	mockService.AssertExpectations(t)
}

func Test_DeleteByUserID_Forbidden(t *testing.T) {
	mockService := new(mocks.MockBlogService)
	validate := validator.New()
	h := NewHandler(mockService, nil, validate)

	userID := uuid.New()
	otherUserID := uuid.New()

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/blogs/user/"+otherUserID.String(), http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(otherUserID.String())
	c.Set("id", userID)
	c.Set("isAdmin", false)

	err := h.DeleteByUserID(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Contains(t, rec.Body.String(), "You need the admin role")

	mockService.AssertExpectations(t)
}

func Test_Update_AsAdmin(t *testing.T) {
	mockService := new(mocks.MockBlogService)
	validate := validator.New()
	h := NewHandler(mockService, nil, validate)

	updBlog := model.Blog{
		BlogID:  uuid.New(),
		Title:   "Updated Title",
		Content: "Updated Content",
	}

	bodyBytes, err := json.Marshal(updBlog)
	require.NoError(t, err)

	mockService.On("Update", mock.Anything, &updBlog).Return(nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/blog", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("isAdmin", true)

	err = h.Update(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	var respBlog model.Blog
	err = json.Unmarshal(rec.Body.Bytes(), &respBlog)
	require.NoError(t, err)
	require.Equal(t, updBlog, respBlog)

	mockService.AssertExpectations(t)
}

func Test_Update_AsUser_OwnBlog(t *testing.T) {
	mockService := new(mocks.MockBlogService)
	validate := validator.New()
	h := NewHandler(mockService, nil, validate)

	userID := uuid.New()
	updBlog := model.Blog{
		BlogID:  uuid.New(),
		Title:   "Updated Title",
		Content: "Updated Content",
	}

	blogs := []*model.Blog{
		{
			BlogID: updBlog.BlogID,
		},
	}

	bodyBytes, err := json.Marshal(updBlog)
	require.NoError(t, err)

	mockService.On("GetByUserID", mock.Anything, userID).Return(blogs, nil)
	mockService.On("Update", mock.Anything, &updBlog).Return(nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/blog", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("id", userID)

	err = h.Update(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	var respBlog model.Blog
	err = json.Unmarshal(rec.Body.Bytes(), &respBlog)
	require.NoError(t, err)
	require.Equal(t, updBlog, respBlog)

	mockService.AssertExpectations(t)
}

func Test_Update_NotOwner(t *testing.T) {
	mockService := new(mocks.MockBlogService)
	validate := validator.New()
	h := NewHandler(mockService, nil, validate)

	userID := uuid.New()
	updBlog := model.Blog{
		BlogID:  uuid.New(),
		Title:   "Updated Title",
		Content: "Updated Content",
	}

	blogs := []*model.Blog{}

	bodyBytes, err := json.Marshal(updBlog)
	require.NoError(t, err)

	mockService.On("GetByUserID", mock.Anything, userID).Return(blogs, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/blog", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("id", userID)

	err = h.Update(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Contains(t, rec.Body.String(), "Cannot update blog with id")

	mockService.AssertExpectations(t)
}

func Test_GetAll(t *testing.T) {
	mockService := new(mocks.MockBlogService)
	validate := validator.New()
	h := NewHandler(mockService, nil, validate)

	blogs := []*model.Blog{
		{BlogID: uuid.New(), Title: "Title1", Content: "Content1"},
		{BlogID: uuid.New(), Title: "Title2", Content: "Content2"},
	}

	mockService.On("GetAll", mock.Anything).Return(blogs, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/blogs", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetAll(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	var respBlogs []*model.Blog
	err = json.Unmarshal(rec.Body.Bytes(), &respBlogs)
	require.NoError(t, err)
	require.Equal(t, blogs, respBlogs)

	mockService.AssertExpectations(t)
}

func Test_GetByUserID(t *testing.T) {
	mockService := new(mocks.MockBlogService)
	validate := validator.New()
	h := NewHandler(mockService, nil, validate)

	userID := uuid.New()
	blogs := []*model.Blog{
		{BlogID: uuid.New(), Title: "Title1", Content: "Content1", UserID: userID},
	}

	mockService.On("GetByUserID", mock.Anything, userID).Return(blogs, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/blogs/user/"+userID.String(), http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(userID.String())
	c.Set("id", userID)

	err := h.GetByUserID(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	var respBlogs []*model.Blog
	err = json.Unmarshal(rec.Body.Bytes(), &respBlogs)
	require.NoError(t, err)
	require.Equal(t, blogs, respBlogs)

	mockService.AssertExpectations(t)
}

func Test_SignUpUser(t *testing.T) {
	mockService := new(mocks.MockUserService)
	validate := validator.New()
	h := NewHandler(nil, mockService, validate)

	inputData := InputData{
		Username: "testuser",
		Password: "password123",
	}
	bodyBytes, err := json.Marshal(inputData)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockService.On("SignUp", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	err = h.SignUpUser(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, rec.Code)
	require.Equal(t, "\"User created\"\n", rec.Body.String())

	mockService.AssertExpectations(t)
}

func Test_SignUpAdmin(t *testing.T) {
	mockService := new(mocks.MockUserService)
	validate := validator.New()
	h := NewHandler(nil, mockService, validate)

	inputData := InputData{
		Username: "adminuser",
		Password: "adminpass",
	}
	bodyBytes, err := json.Marshal(inputData)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/signup/admin", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("isAdmin", true)

	mockService.On("SignUp", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	err = h.SignUpAdmin(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, rec.Code)
	require.Equal(t, "\"Admin created\"\n", rec.Body.String())

	mockService.AssertExpectations(t)
}

func Test_Login(t *testing.T) {
	mockService := new(mocks.MockUserService)
	validate := validator.New()
	h := NewHandler(nil, mockService, validate)

	input := &InputData{
		Username: "testuser",
		Password: "testpassword",
	}

	bodyBytes, err := json.Marshal(input)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	user := &model.User{
		Username: input.Username,
		Password: []byte(input.Password),
	}

	tokenPair := service.TokenPair{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	mockService.On("Login", mock.Anything, user).Return(&tokenPair, nil)

	err = h.Login(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Equal(t, "access-token", response["Access Token : "])
	require.Equal(t, "refresh-token", response["Refresh Token : "])

	mockService.AssertExpectations(t)
}

func Test_Refresh(t *testing.T) {
	mockService := new(mocks.MockUserService)
	validate := validator.New()
	h := NewHandler(nil, mockService, validate)

	input := struct {
		AccessToken  string `json:"accesstoken"`
		RefreshToken string `json:"refreshtoken"`
	}{
		AccessToken:  "oldaccesstoken",
		RefreshToken: "oldrefreshtoken",
	}

	bodyBytes, err := json.Marshal(input)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	updatedTokenPair := service.TokenPair{
		AccessToken:  "newaccesstoken",
		RefreshToken: "newrefreshtoken",
	}

	mockService.On("Refresh", mock.Anything, mock.AnythingOfType("service.TokenPair")).Return(updatedTokenPair, nil)

	err = h.Refresh(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Equal(t, "newaccesstoken", response["Access Token : "])
	require.Equal(t, "newrefreshtoken", response["Refresh Token : "])

	mockService.AssertExpectations(t)
}
