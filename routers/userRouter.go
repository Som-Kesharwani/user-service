package routers

import (
	"github.com/Som-Kesharwani/user-service/controller"
	"github.com/Som-Kesharwani/user-service/middleware"
	"github.com/gin-gonic/gin"
)

func UserRouter(incomingRoutes *gin.Engine) {

	//Test only
	//incomingRoutes.POST("/deleteTokens", controller.DeleteTokens())

	incomingRoutes.POST("/user/signUp", controller.SignUp())
	incomingRoutes.POST("/user/login", controller.Login())
	incomingRoutes.POST("/user/refreshToken", controller.RefreshToken())

	protected := incomingRoutes.Group("/", middleware.Authenticate())
	{
		protected.GET("/getUserToken", controller.GetUserToken())
		protected.GET("/users", controller.GetUsers())
		protected.GET("user/:id", controller.GetUser())
	}
}
