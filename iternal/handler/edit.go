package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/TestTask/iternal/model"
	"github.com/TestTask/iternal/types"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/sirupsen/logrus"
)

func (h *CarHandler) editCarHandle(c *gin.Context) {

	h.log.Info("edit car handler start")

	var car model.Car

	if err := json.NewDecoder(c.Request.Body).Decode(&car); err != nil {
		h.log.Info("cannot decode incoming json")
		h.log.Info("edit car handler failed")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "cannot decode json"})
		return
	}

	car.RegNum = c.Param("regNum")
	if err := h.editCar(car); err != nil {

		h.log.Info("edit car handler failed")

		if errors.Is(err, types.InvalidDataError{}) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
			return
		}

		if errors.Is(err, types.CarNotUpdated{}) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "not found car"})
			return
		}

		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "iternal error"})
		return
	}

	c.Status(200)
}

func (h *CarHandler) editCar(car model.Car) error {

	h.log.Info("edit car start")

	h.log.WithFields(logrus.Fields{
		"message": "edit fields",
		"content": car,
	}).Debug()

	tm, canc := context.WithTimeout(h.ctx, h.OperationTimeout)

	defer canc()

	h.log.Info("send request on change data to db")

	if err := h.carRepo.Edit(tm, car); err != nil {
		h.log.Info("changes to the database were not applied", err)
		h.log.Info("edit car failed")
		return err
	}

	h.log.Info("edit car successful")
	h.log.Info("changes to the database have been applied")

	return nil

}
