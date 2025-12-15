package web

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/cyvqet/connectify/internal/domain"
	"github.com/cyvqet/connectify/internal/repository/cache"
	"github.com/cyvqet/connectify/internal/service"

	"github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserId    int64
	UserEmail string
	UserAgent string
	jwt.RegisteredClaims
}

const (
	emailRegex    = `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	passwordRegex = `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&.])[A-Za-z\d@$!%*?&.]{8,}$`
)

type UserHandler struct {
	svc     service.UserService
	codeSvc service.CodeService
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService) *UserHandler {
	return &UserHandler{
		svc:     svc,
		codeSvc: codeSvc,
	}
}

func (u *UserHandler) RegisterRouter(r *gin.Engine) {
	ug := r.Group("/user")

	ug.POST("/signup", u.Signup)
	ug.POST("/login", u.Login)
	ug.POST("/login_jwt", u.LoginJwt)
	ug.POST("/logout", u.Logout)
	ug.POST("/profile", u.Profile)

	ug.POST("/send_sms_code", u.SendSMSLoginCode)
	ug.POST("/login_sms", u.LoginSMS)
}

func (u *UserHandler) Signup(c *gin.Context) {
	type SignupReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	ok, err := ValidateEmail(req.Email)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "system error"})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "email format error"})
		return
	}

	ok, err = ValidatePassword(req.Password)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "system error"})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "password format error"})
		return
	}

	if req.Password != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"message": "two input passwords are not consistent"})
		return
	}

	err = u.svc.SignUp(c.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if errors.Is(err, service.ErrUserDuplicateEmail) {
		c.JSON(http.StatusOK, gin.H{"message": "email conflict"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "system error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "registration successful"})
}

func (u *UserHandler) Login(c *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	err := u.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvaildUserOrPassword) {
			c.JSON(http.StatusOK, gin.H{"message": "username/password error"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "system error"})
		return
	}

	session := sessions.Default(c)      // Get current request session
	session.Set("userEmail", req.Email) // Store user email in session
	session.Options(sessions.Options{
		MaxAge: 3600,
	})
	err = session.Save() // Save session
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "system error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "login successful"})
}

func (u *UserHandler) LoginJwt(c *gin.Context) {
	type LoginJwtReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginJwtReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	err := u.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvaildUserOrPassword) {
			c.JSON(http.StatusOK, gin.H{"message": "username/password error"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "system error"})
		return
	}

	u.SetJWTToken(c, domain.User{Email: req.Email})

	c.JSON(http.StatusOK, gin.H{"message": "login successful"})
}

func (u *UserHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Options(sessions.Options{
		MaxAge: -1,
	})
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "system error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func (u *UserHandler) Profile(c *gin.Context) {

	type ProfileReq struct {
		Id int64 `json:"id"`
	}

	var req ProfileReq
	if err := c.Bind(&req); err != nil {
		return
	}

	user, err := u.svc.Profile(c.Request.Context(), req.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "system error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    user.Id,
		"email": user.Email,
	})
}

func ValidatePassword(password string) (bool, error) {
	re := regexp2.MustCompile(passwordRegex, 0)
	return re.MatchString(password)
}

func ValidateEmail(email string) (bool, error) {
	re := regexp2.MustCompile(emailRegex, 0)
	return re.MatchString(email)
}

func (u *UserHandler) LoginSMS(c *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	ok, err := u.codeSvc.Verify(c.Request.Context(), "bizLogin", req.Phone, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "system error",
		})
		return
	}
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"message": "verification code is incorrect or expired, please get it again",
		})
		return
	}

	// Find or create user (by phone number)
	user, err := u.svc.FindOrCreate(c.Request.Context(), req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "system error",
		})
		return
	}

	u.SetJWTToken(c, user)

	c.JSON(http.StatusOK, gin.H{"message": "login successful"})
}

func (u *UserHandler) MustGetUserClaims(c *gin.Context) UserClaims {
	// Get user information from claim stored in context by middleware
	claimAny, exists := c.Get("claim")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "system error"})
		c.Abort()
		return UserClaims{}
	}

	claim, ok := claimAny.(UserClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "system error"})
		c.Abort()
		return UserClaims{}
	}

	return claim
}

func (u *UserHandler) SetJWTToken(c *gin.Context, user domain.User) {
	claims := UserClaims{
		UserId:    user.Id,
		UserEmail: user.Email,
		UserAgent: c.Request.UserAgent(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("secret"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "system error"})
		c.Abort()
		return
	}

	c.Header("Jwt-Token", tokenStr)
}

func (u *UserHandler) SendSMSLoginCode(c *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	if req.Phone == "" {
		c.JSON(http.StatusOK, gin.H{
			"message": "please input phone number",
		})
		return
	}

	verificationCode, err := u.codeSvc.Send(c.Request.Context(), "bizLogin", req.Phone)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, gin.H{
			"message": "send successful",
			"code":    verificationCode, // TODO: Remove this line in production environment
		})
	case errors.Is(err, cache.ErrVerificationCodeSendRateLimited):
		c.JSON(http.StatusOK, gin.H{
			"message": "sms send too frequently, please try again later",
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "system error",
		})
	}
}
