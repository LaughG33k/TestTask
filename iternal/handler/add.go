package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/LaughG33k/TestTask/iternal/model"
	"github.com/LaughG33k/TestTask/iternal/types"
	"github.com/LaughG33k/TestTask/pkg/client"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/sirupsen/logrus"
)

func (h *CarHandler) addNewCarsHandle(c *gin.Context) {

	var regs model.AddCarsRequest

	if err := json.NewDecoder(c.Request.Body).Decode(&regs); err != nil {
		h.log.Info("Cannot decoded incoming json")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "cannot decode json"})
		return
	}

	if len(regs.RegNums) == 0 {
		h.log.Info("reg nums is empty")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "regNums can not be empty"})
		return
	}

	if len(regs.RegNums) < 2 {

		if err := h.addNewCar(regs.RegNums[0]); err != nil {

			if errors.Is(err, types.CarNotCreated{}) {
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "car not created. Iternal error"})
				return
			}

			if errors.Is(err, types.CarAlreadyExists{}) {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "car exists"})
				return
			}

			if errors.Is(err, types.InvalidDataError{}) {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "not valid data"})
				return
			}

			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Iternal error"})
			return
		}

		c.IndentedJSON(http.StatusCreated, gin.H{"message": "car added"})
		return

	}

	notAdded, err := h.addNewCars(regs.RegNums...)

	if err != nil {

		if errors.Is(err, types.CarNotCreated{}) {

			h.log.Info("car not create")

			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "car not created. Invalid data", "notAdded": notAdded})
			return
		}

		if errors.Is(err, types.CarAlreadyExists{}) {

			h.log.Info("car exists")

			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "car exists", "notAdded": notAdded})
			return
		}

		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "car not created. Iternal error", "notAdded": notAdded})
		return
	}

	c.IndentedJSON(http.StatusCreated, gin.H{"message": "added", "notAdded": notAdded})

}

func (h *CarHandler) addNewCar(regNum string) error {

	h.log.Info("add new car start")

	h.log.WithFields(logrus.Fields{
		"message": "required reg num",
		"content": regNum,
	}).Debug()

	timeoutCtx, canc := context.WithTimeout(h.ctx, h.OperationTimeout)

	defer canc()

	h.log.Info("send request to api")
	car, err := client.GetInfo(timeoutCtx, "http://127.0.0.1:8081/info", regNum)

	if err != nil {
		h.log.Info("cannot get car from api", err)
		h.log.Info("add new car failed")

		return err
	}

	h.log.Info("request is succeseful")

	h.log.WithFields(logrus.Fields{
		"message":         "get a car from api",
		"contentResponse": car,
	}).Debug()

	h.log.Info("start add a car to db")

	if err := h.carRepo.AddNewCar(timeoutCtx, car); err != nil {
		h.log.Info("failed to add the car to the database", err)
		h.log.Info("add new car failed")
		return err
	}

	h.log.Info("add new car successful")
	h.log.Info("car has been added to db")

	return nil
}

func (h *CarHandler) addNewCars(regNums ...string) ([][]string, error) {

	h.log.Info("start add new cars")

	h.log.WithFields(logrus.Fields{
		"message": "required reg nums",
		"content": regNums,
	}).Debug()

	cars := make([]model.Car, 0, len(regNums))

	notAddedRegNums := make([][]string, 0)

	timeoutCtx, canc := context.WithTimeout(h.ctx, h.OperationTimeout)

	defer canc()

	for _, v := range regNums {

		h.log.Info(fmt.Sprintf("send request to api. Required regNum: %s", v))

		car, err := client.GetInfo(timeoutCtx, "http://127.0.0.1:8081/info", v)

		if err != nil {

			if errors.Is(err, types.BadRequestError{}) {
				h.log.Info(fmt.Sprintf("bad request to api. Required regNum: %s", v))
				notAddedRegNums = append(notAddedRegNums, []string{v, types.BadRequestError{}.Error()})
			} else {
				h.log.Info(fmt.Sprintf("api iternal error. Required regNum: %s", v))
				notAddedRegNums = append(notAddedRegNums, []string{v, "Iternal error"})
			}

			continue
		}

		h.log.WithFields(logrus.Fields{
			"message":         "get a car from api",
			"contentResponse": car,
		}).Debug()

		cars = append(cars, car)

	}

	if len(cars) == 0 {
		return notAddedRegNums, types.IternalServerError{}
	}

	h.log.Info("start adding a cars to db")
	notAdded, err := h.carRepo.AddNewCars(timeoutCtx, cars...)

	notAddedRegNums = append(notAddedRegNums, notAdded...)

	h.log.WithFields(logrus.Fields{
		"message": "not added reg nums",
		"content": notAddedRegNums,
	}).Debug()

	if err != nil {
		h.log.Info("failed to add the cars to the database", err)
		return notAddedRegNums, err
	}

	h.log.Info("cars have been added to the database")

	return notAddedRegNums, nil

}
