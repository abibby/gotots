package main

import (
	"fmt"
	"reflect"

	"github.com/abibby/gotots"
	"github.com/google/osv-scanner/pkg/models"
)

func main() {
	v := models.VulnerabilityResults{}
	t := reflect.TypeOf(v)

	fmt.Print(gotots.GenerateTypes(t))
}
