package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/TestTask/iternal/model"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func makeAnyArr[T any](arr []T) []any {

	res := make([]any, len(arr))

	for _, v := range arr {
		res = append(res, v)
	}

	return res

}

func (h *CarHandler) getPageHandler(c *gin.Context) {

	h.log.Info("get page handler start")

	args := c.Request.URL.Query()

	limit := args.Get("limit")
	pastId := args.Get("pastId")
	periodStart := args.Get("periodStart")
	periodEnd := args.Get("periodEnd")

	periodStartInt := 0
	periodEndInt := 0

	if periodStart != "" {
		res, err := strconv.Atoi(periodStart)

		if err != nil {
			h.log.Info("periodStart is not a number")
			h.log.Info("get page handler failed")
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "periodStart must be a number"})
			return
		}

		periodStartInt = res
	}

	if periodEnd != "" {
		res, err := strconv.Atoi(periodEnd)

		if err != nil {
			h.log.Info("periodEnd is not a number")
			h.log.Info("get page handler failed")
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "periodEnd must be a number"})
			return
		}

		periodEndInt = res
	}

	limitInt := 0

	if limit != "" {
		res, err := strconv.Atoi(limit)

		if err != nil {
			h.log.Info("limit is not a number")
			h.log.Info("get page handler failed")
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "limit must be a number"})
			return
		}

		limitInt = res
	}

	pastIdInt := 0

	if pastId != "" {
		res, err := strconv.Atoi(pastId)

		if err != nil {
			h.log.Info("pastId is not a number")
			h.log.Info("get page handler failed")
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "pastId must be a number"})
			return
		}

		pastIdInt = res
	}

	filter := model.CarFilter{
		Limit:       uint(limitInt),
		PastId:      uint(pastIdInt),
		MarkFilter:  makeAnyArr(args["markFilter"]),
		ModelFilter: makeAnyArr(args["modelFilter"]),
		YearFilter:  makeAnyArr(args["yearFilter"]),
		PeriodStart: periodStartInt,
		PeriodEnd:   periodEndInt,

		PersonFilter: model.Person{
			Name:       args.Get("name"),
			Surname:    args.Get("surname"),
			Patronymic: args.Get("patranomic"),
		},
	}

	cars, err := h.getPage(filter)

	if err != nil {
		h.log.Info("get page handler failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "iternal error"})
		return
	}

	h.log.Info("get page handler successful")

	resp := model.GetPageResponse{Cars: cars}

	c.IndentedJSON(http.StatusOK, resp)
}

func (h *CarHandler) getPage(filter model.CarFilter) ([]model.Car, error) {

	h.log.Info("page receiving start")
	h.log.WithFields(logrus.Fields{
		"messaage": "car filters",
		"content":  filter,
	}).Debug()

	tm, canc := context.WithTimeout(h.ctx, h.OperationTimeout)

	defer canc()

	h.log.Info("a request was sent to the database to receive the page")
	cars, err := h.carRepo.GetPage(tm, filter)

	if err != nil {
		h.log.Info("request to get page from database failed", err)
		h.log.Info("page receiving failed")
		return nil, err
	}

	h.log.Info("request to get page from database completed")
	h.log.WithFields(
		logrus.Fields{
			"message": "retrieved page from database",
			"content": cars,
		},
	).Debug()

	h.log.Info("page receiving successful")

	return cars, nil

}
