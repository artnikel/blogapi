package handler

import (
	"context"
	"net/http"

	"github.com/artnikel/blogapi/internal/model"
	"github.com/artnikel/blogapi/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
)

type BlogService interface {
	Create(ctx context.Context, blog *model.Blog) error
	Get(ctx context.Context, id uuid.UUID) (*model.Blog, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserID(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, blog *model.Blog) error
	GetAll(ctx context.Context) ([]*model.Blog, error)
	GetByUserID(ctx context.Context, id uuid.UUID) ([]*model.Blog, error)
}

type UserService interface {
	SignUp(ctx context.Context, user *model.User) error
	Login(ctx context.Context, user *model.User) (*service.TokenPair, error)
	Refresh(ctx context.Context, tokenPair service.TokenPair) (*service.TokenPair, error)
}

type EntityBlog struct {
	srvBlog  BlogService
	validate *validator.Validate
}

type EntityUser struct {
	srvUser  UserService
	validate *validator.Validate
}

func NewEntityBlog(srvBlog BlogService, validate *validator.Validate) *EntityBlog {
	return &EntityBlog{srvBlog: srvBlog, validate: validate}
}

func NewEntityUser(srvUser UserService, validate *validator.Validate) *EntityUser {
	return &EntityUser{srvUser: srvUser, validate: validate}
}

type InputData struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

func (h *EntityBlog) Create(c echo.Context) error {
	var newBlog model.Blog
	newBlog.BlogID = uuid.New()
	err := c.Bind(&newBlog)
	if err != nil {
		log.Errorf("error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "filling car error")
	}
	userID, ok := c.Get("id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}
	newBlog.UserID = userID
	err = h.validate.StructCtx(c.Request().Context(), newBlog)
	if err != nil {
		log.Errorf("error: %v", err)
		return c.JSON(http.StatusBadRequest, "Not valid data")
	}
	err = h.srvBlog.Create(c.Request().Context(), &newBlog)
	if err != nil {
		log.WithFields(log.Fields{
			"Title":   newBlog.Title,
			"Content": newBlog.Content,
		}).Errorf("srvBlog.Create - %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create blog")
	}
	return c.JSON(http.StatusCreated, newBlog)
}

func (h *EntityBlog) Get(c echo.Context) error {
	id := c.Param("id")
	err := h.validate.VarCtx(c.Request().Context(), id, "required,uuid")
	if err != nil {
		log.Errorf("h.validate.VarCtx error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "failed to validate id")
	}
	uuidID, err := uuid.Parse(id)
	if err != nil {
		log.Errorf("uuid.Parse error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse id")
	}
	blog, err := h.srvBlog.Get(c.Request().Context(), uuidID)
	if err != nil {
		log.WithField("ID", uuidID).Errorf("srvBlog.Get - %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to get blog")
	}
	return c.JSON(http.StatusOK, blog)
}

func (h *EntityBlog) GetAll(c echo.Context) error {
	blogs, err := h.srvBlog.GetAll(c.Request().Context())
	if err != nil {
		log.Errorf("srvBlog.GetAll - %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to get all blogs")
	}
	return c.JSON(http.StatusOK, blogs)
}

func (h *EntityUser) SignUpUser(c echo.Context) error {
	requestData := &InputData{}
	err := c.Bind(requestData)
	if err != nil {
		log.Errorf("error: %v", err)
		return c.JSON(http.StatusBadRequest, "SignUpUser: Invalid request payload")
	}
	newUser := &model.User{
		ID:       uuid.New(),
		Username: requestData.Username,
		Password: []byte(requestData.Password),
		Admin:    false,
	}
	err = h.validate.StructCtx(c.Request().Context(), newUser)
	if err != nil {
		log.Errorf("error: %v", err)
		return c.JSON(http.StatusBadRequest, "Not valid data")
	}
	err = h.srvUser.SignUp(c.Request().Context(), newUser)
	if err != nil {
		log.WithFields(log.Fields{
			"Username": newUser.Username,
			"Password": newUser.Password,
		}).Errorf("srvUser.SignUp - %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to sign up user")
	}
	return c.JSON(http.StatusCreated, "User created")
}

func (h *EntityUser) SignUpAdmin(c echo.Context) error {
	requestData := &InputData{}
	err := c.Bind(requestData)
	if err != nil {
		log.Errorf("error: %v", err)
		return c.JSON(http.StatusBadRequest, "SignUpUser: Invalid request payload")
	}
	newAdmin := &model.User{
		ID:       uuid.New(),
		Username: requestData.Username,
		Password: []byte(requestData.Password),
		Admin:    true,
	}
	err = h.validate.StructCtx(c.Request().Context(), newAdmin)
	if err != nil {
		log.Errorf("error: %v", err)
		return c.JSON(http.StatusBadRequest, "Not valid data")
	}
	err = h.srvUser.SignUp(c.Request().Context(), newAdmin)
	if err != nil {
		log.WithFields(log.Fields{
			"Username": newAdmin.Username,
			"Password": newAdmin.Password,
		}).Errorf("srvUser.SignUpAdmin - %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to sign up admin")
	}
	return c.JSON(http.StatusCreated, "Admin created")
}

func (h *EntityUser) Login(c echo.Context) error {
	requestData := &InputData{}
	err := c.Bind(requestData)
	if err != nil {
		log.Errorf("error: %v", err)
		return c.JSON(http.StatusBadRequest, "SignUpUser: Invalid request payload")
	}
	loginedUser := &model.User{
		Username: requestData.Username,
		Password: []byte(requestData.Password),
	}
	err = h.validate.StructCtx(c.Request().Context(), loginedUser)
	if err != nil {
		log.Errorf("error: %v", err)
		return c.JSON(http.StatusBadRequest, "Not valid data")
	}
	tokenPair, err := h.srvUser.Login(c.Request().Context(), loginedUser)
	if err != nil {
		log.WithFields(log.Fields{
			"Username": loginedUser.Username,
			"Password": loginedUser.Password,
		}).Errorf("srvUser.Login - %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to log in")
	}
	return c.JSON(http.StatusCreated, echo.Map{
		"Access Token : ":  tokenPair.AccessToken,
		"Refresh Token : ": tokenPair.RefreshToken,
	})
}
