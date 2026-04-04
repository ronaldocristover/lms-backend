package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name  string `validate:"required,min=2,max=50"`
	Email string `validate:"required,email"`
	Age   int    `validate:"gte=0,lte=150"`
}

func TestStruct_Valid(t *testing.T) {
	s := TestStruct{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
	}

	err := Struct(s)
	assert.NoError(t, err)
}

func TestStruct_InvalidEmail(t *testing.T) {
	s := TestStruct{
		Name:  "John Doe",
		Email: "invalid-email",
		Age:   25,
	}

	err := Struct(s)
	assert.Error(t, err)
}

func TestStruct_EmptyName(t *testing.T) {
	s := TestStruct{
		Name:  "",
		Email: "john@example.com",
		Age:   25,
	}

	err := Struct(s)
	assert.Error(t, err)
}

func TestStruct_NameTooShort(t *testing.T) {
	s := TestStruct{
		Name:  "J",
		Email: "john@example.com",
		Age:   25,
	}

	err := Struct(s)
	assert.Error(t, err)
}

func TestStruct_AgeNegative(t *testing.T) {
	s := TestStruct{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   -1,
	}

	err := Struct(s)
	assert.Error(t, err)
}

func TestStruct_AgeTooHigh(t *testing.T) {
	s := TestStruct{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   200,
	}

	err := Struct(s)
	assert.Error(t, err)
}

func TestErrors_Format(t *testing.T) {
	s := TestStruct{
		Name:  "",
		Email: "invalid",
		Age:   -1,
	}

	err := Struct(s)
	assert.Error(t, err)

	errors := Errors(err)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "Name")
}
