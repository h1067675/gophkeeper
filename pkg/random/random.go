// Package random создает случайные данные
package random

import (
	"fmt"
	"math/rand"
	"strings"
)

// RandCode создает случайны 6 значный код
func RandCode() string {
	code := fmt.Sprintf("%d", rnd(100, 999999))
	return fmt.Sprintf("%s%s", strings.Repeat("0", 6-len(code)), code)
}

// rnd дает случайное число от mini до maxi
func rnd(mini, maxi int) int {
	res := rand.Intn(maxi-mini) + mini
	return res
}
