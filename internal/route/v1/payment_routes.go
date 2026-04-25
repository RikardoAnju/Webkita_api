package v1

import (
    "BackendFramework/internal/controller"
    "BackendFramework/internal/middleware"
    "github.com/gin-gonic/gin"
)

func PaymentRoutes(r *gin.RouterGroup) {
    payment := r.Group("/payment")
    {
        // Public — Midtrans webhook (tanpa auth)
        payment.POST("/notification", controller.MidtransNotification)

        // Protected — butuh login
        auth := payment.Group("")
        auth.Use(middleware.JWTAuthMiddleware())
        {
            auth.POST("", controller.CreatePayment)              // POST /v1/payment         - buat pembayaran
            auth.GET("/my", controller.GetMyPayments)            // GET  /v1/payment/my      - riwayat pembayaran saya
            auth.GET("/:orderId", controller.GetPaymentByOrder)  // GET  /v1/payment/:orderId - detail pembayaran
        }

        // Admin only
        admin := payment.Group("")
        admin.Use(middleware.JWTAuthMiddleware())
        admin.Use(middleware.AdminMiddleware())
        {
            admin.GET("", controller.GetAllPayments)                       // GET   /v1/payment               - semua pembayaran
            admin.GET("/project/:projectId", controller.GetPaymentsByProject) // GET   /v1/payment/project/:projectId
            admin.PATCH("/:orderId/status", controller.UpdatePaymentStatus) // PATCH /v1/payment/:orderId/status - update manual
        }
    }
}