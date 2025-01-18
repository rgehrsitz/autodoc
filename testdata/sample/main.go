package main

import "./math"

func main() {
    calc := &math.Calculator{}
    result := calc.Add(1, 2)
    println(result)
}