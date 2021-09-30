package notifier

import (
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"

	_ "github.com/alufers/notifier/docs"
)

// @title Notifier by alufers
// @version 1.0
// @description This service accepts notifications as HTTP requests and sends them to various sinks, like Telegram, email etc.
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

type HttpServer struct {
	router *fiber.App
	Sinks  []NotificationSink
	Users  []*User
}

func NewHttpServer(sinks []NotificationSink, users []*User) *HttpServer {
	return &HttpServer{
		router: fiber.New(
			fiber.Config{
				AppName:      "Notifier",
				ServerHeader: "Notifier",
			},
		),
		Sinks: sinks,
		Users: users,
	}
}

func (s *HttpServer) Start(addr string) {

	s.router.Use(s.authorizationMiddleware)
	s.router.Post("/notify", s.postNotify)
	s.router.Post("/question", s.postQuestion)
	s.router.Get("/login", s.getLogin)
	s.router.Post("/login", s.postLogin)
	s.router.Get("/*", swagger.Handler) // default
	if err := s.router.Listen(addr); err != nil {
		log.Fatal(err)
	}
}

func (s *HttpServer) authorizationMiddleware(c *fiber.Ctx) error {
	if string(c.Request().URI().Path()) == "/login" {
		return c.Next()
	}
	tokenString := c.Cookies("NOTIFIER_TOKEN")

	if tokenString != "" {
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(viper.GetString("http.jwt_secret")), nil
		})
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if claims.VerifyExpiresAt(time.Now().Unix(), true) {
				for _, user := range s.Users {
					if user.Username == claims["username"].(string) {
						c.Context().SetUserValue("user", user)
						return c.Next()
					}
				}
			}
		}

	}
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		acceptValue := c.Get("Accept")
		if strings.Contains(acceptValue, "text/html") {
			c.Response().Header.Set("Location", "/login")
			return c.Status(http.StatusSeeOther).SendString("Redirecting to login...")
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header or NOTIFIER_TOKEN cookie is missing or invalid",
		})
	}

	authHeader = strings.TrimPrefix(authHeader, "Bearer ")
	for _, user := range s.Users {
		if user.Token == authHeader {
			c.Context().SetUserValue("user", user)
			return c.Next()
		}
	}

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": "Authorization header is invalid",
	})
}

//go:embed assets/login.html
var loginTemplate []byte

func (s *HttpServer) getLogin(c *fiber.Ctx) error {
	c.Response().Header.Set("Content-Type", "text/html")
	return template.Must(template.New("login").Parse(string(loginTemplate))).Execute(c.Response().BodyWriter(), nil)
}

type PostLoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *HttpServer) postLogin(c *fiber.Ctx) error {
	var body PostLoginBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	log.Printf("Login attempt for user: %v", body.Username)
	for _, user := range s.Users {
		if user.Username == body.Username && user.Password != "" && user.Password == body.Password {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"username": user.Username,
				"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(),
			})
			tokenString, err := token.SignedString([]byte(viper.GetString("http.jwt_secret")))
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": err.Error(),
				})
			}
			c.Cookie(&fiber.Cookie{
				Name:  "NOTIFIER_TOKEN",
				Value: tokenString,
			})
			c.Response().Header.Set("Location", "/")

			return c.Status(http.StatusSeeOther).
				SendString("Logged in successfully, redirecting...")
		}
	}

	c.Response().Header.Set("Content-Type", "text/html")
	return template.Must(template.New("login").Parse(string(loginTemplate))).Execute(c.Response().BodyWriter(), fiber.Map{
		"Message": "Invalid username or password",
	})
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewErrorResponse(err error) *ErrorResponse {
	return &ErrorResponse{
		Error: err.Error(),
	}
}

type PostNotifyResponse struct {
	DeliveriesTotal     int               `json:"deliveriesTotal"`
	DeliveriesSucceeded int               `json:"deliveriesCucceeded"`
	Errors              map[string]string `json:"errors"`
}

type PostNotifyBody struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// postNotify godoc
// @Summary Send a notification
// @Description Delivers a notification to all the sinks
// @ID post-notification
// @Param notification body PostNotifyBody true "Notification to deliver"
// @Accept  json
// @Produce  json
// @Success 200 {object} PostNotifyResponse
// @Failure 400 {object} ErrorResponse
// @Router /notify [post]
// @Security ApiKeyAuth
func (s *HttpServer) postNotify(c *fiber.Ctx) error {
	var body PostNotifyBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	notification := &Notification{
		Timestamp: time.Now(),
		Title:     body.Title,
		Body:      body.Body,
	}
	if notification.Body == "" {
		return c.Status(fiber.StatusBadRequest).JSON(NewErrorResponse(fmt.Errorf("body is empty")))
	}
	var resp PostNotifyResponse
	resp.DeliveriesTotal = len(s.Sinks)
	resp.Errors = make(map[string]string)
	for _, s := range s.Sinks {
		if err := s.DeliverNotification(notification); err != nil {

			log.Printf("Delivery with sink %T failed: %v", s, err)
			resp.Errors[fmt.Sprintf("%T", s)] = err.Error()
		} else {
			resp.DeliveriesSucceeded++
		}
	}
	return c.JSON(resp)
}

type PostQuestionBody struct {
	Text    string        `json:"text"`
	Kind    string        `json:"kind"`
	Timeout time.Duration `json:"timeout" swaggertype:"primitive,string"`
}

type PostQuestionResponse struct {
	Errors map[string]string `json:"errors"`
	Answer *Answer           `json:"answer"`
}

type sinkResult struct {
	sinkName string
	err      error
	answer   *Answer
}

// postQuestion godoc
// @Summary Asks a question to the user
// @Description Currently supported question types: yesno
// @ID post-question
// @Param notification body PostQuestionBody true "Question to ask"
// @Accept  json
// @Produce  json
// @Success 200 {object} PostQuestionResponse
// @Failure 400 {object} ErrorResponse
// @Router /question [post]
// @Security ApiKeyAuth
func (s *HttpServer) postQuestion(c *fiber.Ctx) error {
	var body PostQuestionBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if body.Kind == "" {
		body.Kind = string(QuestionKind_YesNo)
	}
	question := &Question{
		Timestamp: time.Now(),
		Text:      body.Text,
		Kind:      QuestionKind(body.Kind),
	}
	if question.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "text is empty",
		})
	}

	if body.Timeout < time.Second {
		body.Timeout = time.Hour * 100000
	}

	ctx, cancel := context.WithTimeout(c.Context(), body.Timeout)
	go func() {
		<-c.Context().Done()
		log.Printf("THE CTX IS DONE NOW")
	}()
	resultsChan := make(chan sinkResult, len(s.Sinks))
	var totalSinksAsked int
	for _, sink := range s.Sinks {
		if sinkWithQuestions, ok := sink.(NotificationSinkWithQuestions); ok {
			totalSinksAsked++
			go func() {
				ans, err := sinkWithQuestions.AskQuestion(ctx, question)
				if err != nil {
					resultsChan <- sinkResult{sinkName: fmt.Sprintf("%T", sink), err: err}
				}
				resultsChan <- sinkResult{sinkName: fmt.Sprintf("%T", sink), answer: ans}
			}()
		}
	}

	errorsMap := make(map[string]string)
	var answer *Answer
	var i int
	for result := range resultsChan {
		i++
		if result.err != nil {
			errorsMap[result.sinkName] = result.err.Error()
		} else {
			answer = result.answer

			break
		}
		if i == totalSinksAsked {
			break
		}
	}
	cancel()

	return c.JSON(&PostQuestionResponse{
		Errors: errorsMap,
		Answer: answer,
	})
}
