/*
			NON Trivial Example
		Higher Order Functions - Go lang
*/
package main
import "fmt"

//simple function
func sum(x, y int) int {
	return x + y
}

//higher order function
// arg : int, return : a func
func partialSum(x int) func(int) int {
	return func(y int) int {
		return sum(x, y)
	}
}
func main() {
	partial := partialSum(3)
	fmt.Println(partial(7))
}