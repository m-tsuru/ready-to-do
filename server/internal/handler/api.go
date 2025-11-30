package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/m-tsuru/ready-to-do/internal/tasks"
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

	tasks := v1.Group("/tasks")
	tasks.Post("/", createTaskHandler(db))
	tasks.Get("/:id", getTaskHandler(db))
	tasks.Get("/:id/parents", getParentTasksHandler(db))
	tasks.Get("/:id/children", getChildTasksHandler(db))
	tasks.Post("/:id/running", makeTaskRunningHandler(db))
	tasks.Post("/:id/waiting", makeTaskWaitingHandler(db))
	tasks.Post("/:id/done", makeTaskDoneHandler(db))
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

		existingUser, _ := users.GetUserByMail(db, req.Mail)
		if existingUser != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Email already registered",
			})
		}

		user, err := users.CreateUser(db, req.Mail, req.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create user",
			})
		}

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

		user, err := users.GetUserByMail(db, req.Mail)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		valid, err := user.CheckPassword(req.Password)
		if err != nil || !valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

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

		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		_, session, err := users.GetUserSessionByToken(db, token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

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

		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

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

func getUserFromToken(c *fiber.Ctx, db *gorm.DB) (*users.User, error) {
	token := c.Get("Authorization")
	if token == "" {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "No token provided")
	}

	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	user, _, err := users.GetUserSessionByToken(db, token)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}

	return user, nil
}

type CreateTaskRequest struct {
	Name          string   `json:"name" validate:"required"`
	Description   string   `json:"description"`
	RelatedUrl    string   `json:"related_url"`
	ParentTaskIds []string `json:"parent_task_ids"`
}

func createTaskHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, err := getUserFromToken(c, db)
		if err != nil {
			return err
		}

		var req CreateTaskRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		task, err := tasks.CreateTask(db, user.Id, req.Name, req.Description, req.RelatedUrl, req.ParentTaskIds)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create task",
			})
		}

		state, _ := task.GetCurrentState(db)
		isReady, _ := task.IsReady(db)

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"task": fiber.Map{
				"id":          task.Id,
				"name":        task.Name,
				"description": task.Description,
				"related_url": task.RelatedUrl,
				"user_id":     task.UserId,
				"created_at":  task.CreatedAt,
				"state":       state.State,
				"is_ready":    isReady,
			},
		})
	}
}

func getTaskHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, err := getUserFromToken(c, db)
		if err != nil {
			return err
		}

		taskId := c.Params("id")
		if taskId == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Task ID is required",
			})
		}

		task, err := tasks.GetTaskById(db, taskId)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Task not found",
			})
		}

		if task.UserId != user.Id {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}
		state, _ := task.GetCurrentState(db)
		isReady, _ := task.IsReady(db)

		return c.JSON(fiber.Map{
			"task": fiber.Map{
				"id":          task.Id,
				"name":        task.Name,
				"description": task.Description,
				"related_url": task.RelatedUrl,
				"user_id":     task.UserId,
				"created_at":  task.CreatedAt,
				"state":       state.State,
				"is_ready":    isReady,
			},
		})
	}
}

func getParentTasksHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, err := getUserFromToken(c, db)
		if err != nil {
			return err
		}

		taskId := c.Params("id")
		if taskId == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Task ID is required",
			})
		}

		task, err := tasks.GetTaskById(db, taskId)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Task not found",
			})
		}

		if task.UserId != user.Id {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		parentTasks, err := tasks.GetParentDependency(db, task)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get parent tasks",
			})
		}

		taskList := make([]fiber.Map, 0)
		for _, pt := range *parentTasks {
			state, _ := pt.GetCurrentState(db)
			isReady, _ := pt.IsReady(db)
			taskList = append(taskList, fiber.Map{
				"id":          pt.Id,
				"name":        pt.Name,
				"description": pt.Description,
				"related_url": pt.RelatedUrl,
				"state":       state.State,
				"is_ready":    isReady,
			})
		}

		return c.JSON(fiber.Map{
			"parents": taskList,
		})
	}
}

func getChildTasksHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, err := getUserFromToken(c, db)
		if err != nil {
			return err
		}

		taskId := c.Params("id")
		if taskId == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Task ID is required",
			})
		}

		task, err := tasks.GetTaskById(db, taskId)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Task not found",
			})
		}

		if task.UserId != user.Id {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		childTasks, err := tasks.GetChildDependency(db, task)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get child tasks",
			})
		}

		taskList := make([]fiber.Map, 0)
		for _, ct := range *childTasks {
			state, _ := ct.GetCurrentState(db)
			isReady, _ := ct.IsReady(db)
			taskList = append(taskList, fiber.Map{
				"id":          ct.Id,
				"name":        ct.Name,
				"description": ct.Description,
				"related_url": ct.RelatedUrl,
				"state":       state.State,
				"is_ready":    isReady,
			})
		}

		return c.JSON(fiber.Map{
			"children": taskList,
		})
	}
}

func makeTaskRunningHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, err := getUserFromToken(c, db)
		if err != nil {
			return err
		}

		taskId := c.Params("id")
		if taskId == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Task ID is required",
			})
		}

		task, err := tasks.GetTaskById(db, taskId)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Task not found",
			})
		}

		if task.UserId != user.Id {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		success, err := task.MakeRunning(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update task state",
			})
		}

		if !success {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Task is not ready to run",
			})
		}

		state, _ := task.GetCurrentState(db)

		return c.JSON(fiber.Map{
			"message": "Task state updated to running",
			"task": fiber.Map{
				"id":    task.Id,
				"state": state.State,
			},
		})
	}
}

func makeTaskWaitingHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, err := getUserFromToken(c, db)
		if err != nil {
			return err
		}

		taskId := c.Params("id")
		if taskId == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Task ID is required",
			})
		}

		task, err := tasks.GetTaskById(db, taskId)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Task not found",
			})
		}

		if task.UserId != user.Id {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		success, err := task.MakeWaiting(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update task state",
			})
		}

		if !success {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Task is not running",
			})
		}

		state, _ := task.GetCurrentState(db)

		return c.JSON(fiber.Map{
			"message": "Task state updated to waiting",
			"task": fiber.Map{
				"id":    task.Id,
				"state": state.State,
			},
		})
	}
}

func makeTaskDoneHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, err := getUserFromToken(c, db)
		if err != nil {
			return err
		}

		taskId := c.Params("id")
		if taskId == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Task ID is required",
			})
		}

		task, err := tasks.GetTaskById(db, taskId)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Task not found",
			})
		}

		if task.UserId != user.Id {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		success, err := task.MakeDone(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update task state",
			})
		}

		if !success {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Task cannot be marked as done",
			})
		}

		state, _ := task.GetCurrentState(db)

		return c.JSON(fiber.Map{
			"message": "Task state updated to done",
			"task": fiber.Map{
				"id":    task.Id,
				"state": state.State,
			},
		})
	}
}
