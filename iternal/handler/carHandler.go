package handler

import (
	"context"
	"time"

	"github.com/TestTask/iternal/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type CarHandler struct {
	ctx context.Context

	log              *logrus.Logger
	router           *gin.Engine
	OperationTimeout time.Duration
	carRepo          repository.CarRepository
}

func NewCarHandler(log *logrus.Logger, router *gin.Engine, carRepo repository.CarRepository) *CarHandler {
	return &CarHandler{
		ctx:     context.Background(),
		log:     log,
		router:  router,
		carRepo: carRepo,
	}
}

func (h *CarHandler) Start() {
	h.router.POST("/addCars", h.addNewCarsHandle)
	h.router.PATCH("/editCar/:regNum", h.editCarHandle)
	h.router.GET("/getPage", h.getPageHandler)
	h.router.DELETE("/deleteCar/:regNum", h.deleteHandle)
}
