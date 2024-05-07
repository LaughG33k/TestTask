package types

type InvalidUuidError struct{}

func (InvalidUuidError) Error() string {
	return "Invalid uuid"
}

type InvalidDataError struct{}

func (InvalidDataError) Error() string {
	return "Invalid data"
}

type CarAlreadyExists struct{}

func (CarAlreadyExists) Error() string {
	return "car already exists"
}

type CarNoExist struct{}

func (CarNoExist) Error() string {
	return "car no exists"
}

type CarNotCreated struct{}

func (CarNotCreated) Error() string {
	return "car not created"
}

type CarNotUpdated struct{}

func (CarNotUpdated) Error() string {
	return "car not updated"
}

type BadRequestError struct{}

func (BadRequestError) Error() string {
	return "Bad request"
}

type IternalServerError struct{}

func (IternalServerError) Error() string {
	return "Iternal server error"
}
