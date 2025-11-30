package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/m-tsuru/ready-to-do/internal/users"
	"gorm.io/gorm"
)

func Handler(app *fiber.App, db *gorm.DB) {
	v1 := app.Group("/api/v1")
	v1.Get("/", pingHandler)

	auth := v1.Group("/auth")
	auth.Post("/register", registerHandler(db))
	auth.Post("/login", loginHandler(db))
	auth.Post("/logout", logoutHandler(db))
	auth.Get("/me", meHandler(db))
}

func pingHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "pong",
		"status":  "ok",
	})
}

type RegisterRequest struct {
	Mail     string `json:"mail" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func registerHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RegisterRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// メールアドレスの重複チェック
		existingUser, _ := users.GetUserByMail(db, req.Mail)
		if existingUser != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Email already registered",
			})
		}

		// ユーザー作成
		user, err := users.CreateUser(db, req.Mail, req.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create user",
			})
		}

		// セッション作成
		session, err := user.CreateUserSession(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create session",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"user": fiber.Map{
				"id":   user.Id,
				"mail": user.Mail,
			},
			"token": session.Token,
		})
	}
}

type LoginRequest struct {
	Mail     string `json:"mail" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func loginHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// ユーザー取得
		user, err := users.GetUserByMail(db, req.Mail)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		// パスワード検証
		valid, err := user.CheckPassword(req.Password)
		if err != nil || !valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		// セッション作成
		session, err := user.CreateUserSession(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create session",
			})
		}

		return c.JSON(fiber.Map{
			"user": fiber.Map{
				"id":   user.Id,
				"mail": user.Mail,
			},
			"token": session.Token,
		})
	}
}

func logoutHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "No token provided",
			})
		}

		// "Bearer " プレフィックスを削除
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// セッション取得
		_, session, err := users.GetUserSessionByToken(db, token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// セッション削除
		if err := db.Delete(session).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to logout",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Logged out successfully",
		})
	}
}

func meHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "No token provided",
			})
		}

		// "Bearer " プレフィックスを削除
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// セッションとユーザー取得
		user, _, err := users.GetUserSessionByToken(db, token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		return c.JSON(fiber.Map{
			"user": fiber.Map{
				"id":         user.Id,
				"mail":       user.Mail,
				"created_at": user.CreatedAt,
				"updated_at": user.UpdatedAt,
			},
		})
	}
}
