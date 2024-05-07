package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/LaughG33k/TestTask/iternal/model"
	"github.com/LaughG33k/TestTask/iternal/types"

	"github.com/goccy/go-json"
)

func GetInfo(ctx context.Context, path, regNum string) (model.Car, error) {

	client := &http.Client{}

	url := fmt.Sprintf("%s/info?regNum=%s", os.Getenv("CAR_INFO_API_URL"), regNum)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	var car model.Car

	if err != nil {
		return car, err
	}

	resp, err := client.Do(req)

	if err != nil {
		return car, err
	}

	if resp.StatusCode != 200 {

		if resp.StatusCode == 400 {
			return car, types.BadRequestError{}
		}

		if resp.StatusCode == 500 {
			return car, types.IternalServerError{}
		}

		return car, errors.New(resp.Status)
	}

	bytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return car, err
	}

	if err := json.Unmarshal(bytes, &car); err != nil {
		return car, err
	}

	return car, nil
}
