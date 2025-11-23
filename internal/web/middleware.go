package web

import (
	"log"
	"time"

	"github.com/labstack/echo/v4"
)

func AccessLogMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		err := next(c)

		stop := time.Since(start)
		req := c.Request()
		res := c.Response()

		log.Printf("%s %s %d %s", req.Method, req.URL.Path, res.Status, stop)

		return err
	}
}
