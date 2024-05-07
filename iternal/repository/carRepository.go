package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/testTask/iternal/model"
	"github.com/testTask/iternal/types"
	"github.com/testTask/pkg/client/psql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type CarRepository interface {
	AddNewCar(ctx context.Context, car model.Car) error
	AddNewCars(ctx context.Context, cars ...model.Car) ([][]string, error)
	Edit(ctx context.Context, car model.Car) error
	Delete(ctx context.Context, regNum string) error
	GetPage(ctx context.Context, filter model.CarFilter) ([]model.Car, error)
}

type carRepo struct {
	client psql.Client
}

func NewCarRepository(client psql.Client) CarRepository {

	return &carRepo{

		client: client,
	}

}

// возвращает страницу данных из бд, которая была отфильтрована.
func (c *carRepo) GetPage(ctx context.Context, filter model.CarFilter) ([]model.Car, error) {

	res := make([]model.Car, 0, filter.Limit)

	if len(filter.YearFilter) > 0 {
		filter.PeriodStart = 0
		filter.PeriodEnd = 0
	}

	query, args := carGPQuery(filter.Limit, filter.PastId, filter.MarkFilter, filter.ModelFilter, filter.YearFilter, filter.PeriodStart, filter.PeriodEnd, filter.PersonFilter)

	rows, err := c.client.Query(ctx, query, args...)

	if err != nil {
		return []model.Car{}, err
	}

	defer rows.Close()

	for rows.Next() {

		var car model.Car

		if err := rows.Scan(&car.Id, &car.RegNum, &car.Mark, &car.Model, &car.Year, &car.Name, &car.Surname, &car.Patronymic); err != nil {
			return []model.Car{}, err
		}

		res = append(res, car)

	}

	return res, nil
}

func (c *carRepo) AddNewCar(ctx context.Context, car model.Car) error {

	if !carValidation(car) {
		return types.InvalidDataError{}
	}

	car.RegNum = strings.ToUpper(car.RegNum)

	query := "insert into cars(reg_num, mark, model, year, owner_name, owner_surname, owner_patranomic) values ($1, $2, $3, $4, $5, $6, $7);"
	args := []any{car.RegNum, car.Mark, car.Model, car.Year, car.Name, car.Surname, car.Patronymic}

	res, err := c.client.Exec(ctx, query, args...)

	if err != nil {

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {

			if pgErr.Code == "23505" {
				return types.CarAlreadyExists{}
			}
		}

		return err
	}

	if res.RowsAffected() == 0 {
		return types.CarNotCreated{}
	}

	return nil
}

// Эта функция добовляет за один запрос все валидные автомобили, которые передали в аргумент.
// Так же функция возвращает 2 массива. Первый содержит список номеров, которые уже
// имеются в базе данных. Второй содержит не добавленые автомобили из-за не прохождение валидации
func (c *carRepo) AddNewCars(ctx context.Context, cars ...model.Car) ([][]string, error) {

	notAdded := make([][]string, 0)

	query := "select reg_num from cars where reg_num in ("

	args := make([]any, 0, len(cars))

	validCars := make([]model.Car, 0, len(cars))

	index := 0

	for _, car := range cars {

		if !carValidation(car) {
			notAdded = append(notAdded, []string{car.RegNum, types.InvalidDataError{}.Error()})
			continue
		}

		car.RegNum = strings.ToUpper(car.RegNum)

		validCars = append(validCars, car)
		args = append(args, car.RegNum)
		query += fmt.Sprintf("$%d,", index+1)

		index++

	}

	if len(validCars) == 0 {
		return notAdded, types.CarNotCreated{}
	}

	query = query[:len(query)-1]

	query += ");"

	rows, err := c.client.Query(ctx, query, args...)

	if err != nil {
		return [][]string{}, err
	}

	defer rows.Close()

	dExisted := make(map[string]struct{})

	for rows.Next() {

		var exRegNum string

		if err := rows.Scan(&exRegNum); err != nil {
			return [][]string{}, err
		}

		notAdded = append(notAdded, []string{exRegNum, types.CarAlreadyExists{}.Error()})
		dExisted[exRegNum] = struct{}{}
	}

	validCarRows := make([][]any, 0)

	for _, car := range validCars {
		if _, ok := dExisted[car.RegNum]; !ok {

			args := []any{
				car.RegNum, car.Mark, car.Model, car.Year, car.Name, car.Surname, car.Patronymic,
			}

			validCarRows = append(validCarRows, args)
		}
	}

	if len(validCarRows) == 0 {
		return notAdded, types.CarNotCreated{}
	}

	if _, err := c.client.CopyFrom(ctx, pgx.Identifier{"cars"}, []string{"reg_num", "mark", "model", "year", "owner_name", "owner_surname", "owner_patranomic"}, pgx.CopyFromRows(validCarRows)); err != nil {

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {

			if pgErr.Code == "23505" {
				return [][]string{}, types.CarAlreadyExists{}
			}
		}

		return [][]string{}, err
	}

	return notAdded, nil
}

func (c *carRepo) Delete(ctx context.Context, regNum string) error {

	query := "delete from cars where reg_num = $1;"

	res, err := c.client.Exec(ctx, query, regNum)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return types.CarNoExist{}
	}

	return nil
}

func (c *carRepo) Edit(ctx context.Context, car model.Car) error {

	if !carValidation(car) {
		return types.InvalidDataError{}
	}

	editFields := make([][]any, 0)

	if car.Mark != "" {
		editFields = append(editFields, []any{"mark", car.Mark})
	}

	if car.Model != "" {
		editFields = append(editFields, []any{"model", car.Model})
	}

	if car.Year > 0 {
		editFields = append(editFields, []any{"year", car.Year})
	}

	if car.Name != "" {
		editFields = append(editFields, []any{"owner_name", car.Name})
	}

	if car.Surname != "" {
		editFields = append(editFields, []any{"owner_surname", car.Surname})
	}

	if car.Patronymic != "" {
		editFields = append(editFields, []any{"owner_patranomic", car.Patronymic})
	}

	if err := c.edit(ctx, "cars", "reg_num", car.RegNum, editFields); err != nil {
		return err
	}

	return nil
}

func (c *carRepo) edit(ctx context.Context, tableName string, indificator string, indificatorVal any, editFields [][]any) error {

	if len(editFields) == 0 {
		return types.InvalidDataError{}
	}

	firstParam := true

	query := fmt.Sprintf("update %s set", tableName)
	args := make([]any, len(editFields))

	for i, v := range editFields {

		args[i] = v[1]

		if firstParam {

			query += fmt.Sprintf(" %v = $%d", v[0], i+1)
			firstParam = false

			continue
		}

		query += fmt.Sprintf(", %v = $%d", v[0], i+1)

	}

	query += fmt.Sprintf(" where %s = $%d;", indificator, len(editFields)+1)
	args = append(args, indificatorVal)

	res, err := c.client.Exec(ctx, query, args...)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return types.CarNotUpdated{}
	}

	return nil
}

func carValidation(car model.Car) bool {

	if len(car.RegNum) > 100 || len(car.RegNum) == 0 {
		return false
	}

	if len(car.Model) > 100 || len(car.Mark) > 100 || len(car.Name) > 100 || len(car.Surname) > 100 || len(car.Patronymic) > 100 {
		return false
	}

	return true

}

func carGPQuery(limit, pastId uint, filterByMark, filterByModel, filterByYear []any, periodStart, periodEnd int, filterByPerson model.Person) (string, []any) {

	fristFilter := true
	indexArgument := 1
	args := make([]any, 0)
	query := "select id, reg_num, mark, model, year, owner_name, owner_surname, owner_patranomic from cars where"

	if pastId > 0 {
		query += fmt.Sprintf(" id > $%d", indexArgument)
		indexArgument++
		args = append(args, pastId)
		fristFilter = false
	}

	concIfFirst := func(in string) string {
		if fristFilter {
			fristFilter = false
			return " " + in
		}
		return " and " + in
	}

	f := func(b bool, returnVal any) any {
		if b {
			return returnVal
		}
		return nil
	}

	filters := [][]any{
		[]any{"mark in(%s)", f(len(filterByMark) > 0, filterByMark)},
		[]any{"model in(%s)", f(len(filterByModel) > 0, filterByModel)},
		[]any{"year in(%s)", f(len(filterByYear) > 0, filterByYear)},
		[]any{"year >= %s", f(periodStart > 0, periodStart)},
		[]any{"year <= %s", f(periodEnd > 0, periodEnd)},
		[]any{"owner_name = %s", f(filterByPerson.Name != "", filterByPerson.Name)},
		[]any{"owner_surname = %s", f(filterByPerson.Surname != "", filterByPerson.Surname)},
		[]any{"owner_patranomic = %s", f(filterByPerson.Patronymic != "", filterByPerson.Patronymic)},
	}

	for _, v := range filters {

		if v[1] == nil {
			continue
		}

		if arr, ok := v[1].([]any); ok {

			tmp := ""
			for i := 0; i < len(arr); i++ {
				tmp += fmt.Sprintf(",$%d", indexArgument)
				indexArgument++
				args = append(args, arr[i])
			}

			tmp = tmp[1:]

			v[0] = concIfFirst(v[0].(string))
			query += fmt.Sprintf(v[0].(string), tmp)
			continue

		}

		args = append(args, v[1])

		v[0] = concIfFirst(v[0].(string))
		query += fmt.Sprintf(v[0].(string), fmt.Sprintf("$%d", indexArgument))
		indexArgument++

	}

	if len(args) == 0 {
		args = append(args, limit)
		return fmt.Sprintf("%s order by id limit $%d", "select id, reg_num, mark, model, year, owner_name, owner_surname, owner_patranomic from cars", indexArgument), args
	}

	args = append(args, limit)

	return fmt.Sprintf("%s order by id limit $%d", query, indexArgument), args

}
