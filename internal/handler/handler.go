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
	Refresh(ctx context.Context, tokenPair service.TokenPair) (service.TokenPair, error)
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
		log.Errorf("c.Bind error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Filling blog error")
	}
	userID, ok := c.Get("id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}
	newBlog.UserID = userID
	err = h.validate.StructCtx(c.Request().Context(), newBlog)
	if err != nil {
		log.Errorf("validate.StructCtx error: %v", err)
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
		log.Errorf("validate.VarCtx error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to validate id")
	}
	uuidID, err := uuid.Parse(id)
	if err != nil {
		log.Errorf("uuid.Parse error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to parse id")
	}
	blog, err := h.srvBlog.Get(c.Request().Context(), uuidID)
	if err != nil {
		log.WithField("ID", uuidID).Errorf("srvBlog.Get - %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to get blog")
	}
	return c.JSON(http.StatusOK, blog)
}

func (h *EntityBlog) Delete(c echo.Context) error {
	id := c.Param("id")
	err := h.validate.VarCtx(c.Request().Context(), id, "required,uuid")
	if err != nil {
		log.Errorf("validate.VarCtx error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to validate id")
	}
	uuidID, err := uuid.Parse(id)
	if err != nil {
		log.Errorf("uuid.Parse error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to parse id")
	}
	isAdmin, ok := c.Get("isAdmin").(bool)
	if ok && isAdmin {
		err = h.srvBlog.Delete(c.Request().Context(), uuidID)
		if err != nil {
			log.WithField("ID", uuidID).Errorf("srvBlog.Delete - %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to delete blog")
		}
		return c.JSON(http.StatusOK, "Deleted: "+id)
	}
	userID, ok := c.Get("id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}
	blogs, err := h.srvBlog.GetByUserID(c.Request().Context(), userID)
	if err != nil {
		log.Errorf("srvBlog.GetByUserID - %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to get blogs by user id")
	}
	for _, blog := range blogs {
		if uuidID == blog.BlogID {
			err = h.srvBlog.Delete(c.Request().Context(), uuidID)
			if err != nil {
				log.WithField("ID", uuidID).Errorf("srvBlog.Delete - %v", err)
				return echo.NewHTTPError(http.StatusBadRequest, "Failed to delete blog")
			}
			return c.JSON(http.StatusOK, "Deleted: "+id)
		}
	}
	return c.JSON(http.StatusNotFound, "Cannot delete blog with id: "+id)
}

func (h *EntityBlog) DeleteByUserID(c echo.Context) error {
	id := c.Param("id")
	err := h.validate.VarCtx(c.Request().Context(), id, "required,uuid")
	if err != nil {
		log.Errorf("validate.VarCtx error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to validate id")
	}
	uuidID, err := uuid.Parse(id)
	if err != nil {
		log.Errorf("uuid.Parse error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to parse id")
	}
	userID, ok := c.Get("id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}
	if userID != uuidID {
		isAdmin, ok := c.Get("isAdmin").(bool)
		if !ok || !isAdmin {
			return c.JSON(http.StatusForbidden, "You need the admin role to delete someone else's blog")
		}
	}
	err = h.srvBlog.DeleteByUserID(c.Request().Context(), userID)
	if err != nil {
		log.WithField("ID", userID).Errorf("srvBlog.DeleteByUserID - %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to delete blogs")
	}
	return c.JSON(http.StatusOK, "Deleted from user id: "+userID.String())
}

func (h *EntityBlog) Update(c echo.Context) error {
	var updBlog model.Blog
	err := c.Bind(&updBlog)
	if err != nil {
		log.Errorf("c.Bind error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Filling blog error")
	}
	err = h.validate.StructCtx(c.Request().Context(), updBlog)
	if err != nil {
		log.Errorf("validate.StructCtx error: %v", err)
		return c.JSON(http.StatusBadRequest, "Not valid data")
	}
	isAdmin, ok := c.Get("isAdmin").(bool)
	if ok && isAdmin {
		err = h.srvBlog.Update(c.Request().Context(), &updBlog)
		if err != nil {
			log.WithFields(log.Fields{
				"Title":   updBlog.Title,
				"Content": updBlog.Content,
			}).Errorf("srvBlog.Update - %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update blog")
		}
		return c.JSON(http.StatusOK, updBlog)
	}
	userID, ok := c.Get("id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}
	blogs, err := h.srvBlog.GetByUserID(c.Request().Context(), userID)
	if err != nil {
		log.Errorf("srvBlog.GetByUserID - %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to get blogs by user id")
	}
	for _, blog := range blogs {
		if updBlog.BlogID == blog.BlogID {
			err = h.srvBlog.Update(c.Request().Context(), &updBlog)
			if err != nil {
				log.WithFields(log.Fields{
					"Title":   updBlog.Title,
					"Content": updBlog.Content,
				}).Errorf("srvBlog.Update - %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update blog")
			}
			return c.JSON(http.StatusOK, updBlog)
		}
	}
	return c.JSON(http.StatusNotFound, "Cannot update blog with id: "+updBlog.BlogID.String())
}

func (h *EntityBlog) GetAll(c echo.Context) error {
	blogs, err := h.srvBlog.GetAll(c.Request().Context())
	if err != nil {
		log.Errorf("srvBlog.GetAll - %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to get all blogs")
	}
	return c.JSON(http.StatusOK, blogs)
}

func (h *EntityBlog) GetByUserID(c echo.Context) error {
	userID, ok := c.Get("id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}
	blogs, err := h.srvBlog.GetByUserID(c.Request().Context(), userID)
	if err != nil {
		log.Errorf("srvBlog.GetByUserID - %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to get blogs by user id")
	}
	return c.JSON(http.StatusOK, blogs)
}

func (h *EntityUser) SignUpUser(c echo.Context) error {
	requestData := &InputData{}
	err := c.Bind(requestData)
	if err != nil {
		log.Errorf("c.Bind error: %v", err)
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
		log.Errorf("validate.StructCtx error: %v", err)
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
	isAdmin, ok := c.Get("isAdmin").(bool)
	if !ok || !isAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, "Admin role not found in context")
	}
	requestData := &InputData{}
	err := c.Bind(requestData)
	if err != nil {
		log.Errorf("c.Bind error: %v", err)
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
		log.Errorf("validate.StructCtx error: %v", err)
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
		log.Errorf("c.Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, "SignUpUser: Invalid request payload")
	}
	loginedUser := &model.User{
		Username: requestData.Username,
		Password: []byte(requestData.Password),
	}
	err = h.validate.StructCtx(c.Request().Context(), loginedUser)
	if err != nil {
		log.Errorf("validate.StructCtx error: %v", err)
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

func (h *EntityUser) Refresh(c echo.Context) error {
	bindInfo := struct {
		AccessToken  string `json:"accesstoken"`
		RefreshToken string `json:"refreshtoken"`
	}{}
	err := c.Bind(&bindInfo)
	if err != nil {
		log.Errorf("c.Bind error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to bind tokens")
	}
	var tokenPair service.TokenPair
	tokenPair.AccessToken = bindInfo.AccessToken
	tokenPair.RefreshToken = bindInfo.RefreshToken
	tokenPair, err = h.srvUser.Refresh(c.Request().Context(), tokenPair)
	if err != nil {
		log.WithFields(log.Fields{
			"AccessToken":  tokenPair.AccessToken,
			"RefreshToken": tokenPair.RefreshToken,
		}).Errorf("srvUser.Refresh - %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to refresh tokens")
	}
	return c.JSON(http.StatusOK, echo.Map{
		"Access Token : ":  tokenPair.AccessToken,
		"Refresh Token : ": tokenPair.RefreshToken,
	})
}
