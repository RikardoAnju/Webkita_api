package v1

import (
	"github.com/gin-gonic/gin"
)

// InitRoutes menginisialisasi semua rute untuk API v1
func InitRoutes(r *gin.RouterGroup) {
	InitAuthRoutes(r)
	InitUserRoutes(r)
	InitProjectRoutes(r)
	InitMiscRoutes(r)
}