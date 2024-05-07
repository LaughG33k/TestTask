package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/LaughG33k/TestTask/iternal/types"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *CarHandler) deleteHandle(c *gin.Context) {

	h.log.Info("start delete handler")

	regNum := c.Param("regNum")

	h.log.WithFields(logrus.Fields{
		"message": "required regNum",
		"content": regNum,
	}).Debug()

	if err := h.delete(regNum); err != nil {

		h.log.Info("delete handler failed")

		if errors.Is(err, types.CarNoExist{}) {
			h.log.Info("car not found")
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "car not found"})
			return
		}

		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "iternal error"})
		return

	}

	h.log.Info("delete handler successful")
	c.Status(200)
}

func (h *CarHandler) delete(regNum string) error {

	h.log.Info("start delete")

	tm, canc := context.WithTimeout(h.ctx, h.OperationTimeout)

	defer canc()

	h.log.Info("request to delete the car has been sent")

	if err := h.carRepo.Delete(tm, regNum); err != nil {
		h.log.Info("request to delete the car failed")
		return err
	}

	h.log.Info("request to delete machine successful")

	return nil
}
