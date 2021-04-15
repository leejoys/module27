package main

import (
	"bufio"
	"container/ring"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

//Напишите код, реализующий пайплайн, работающий с целыми числами и состоящий из следующих стадий:

//[+]
//Стадия фильтрации отрицательных чисел (не пропускать отрицательные числа).
//[+]
//Стадия фильтрации чисел, не кратных 3 (не пропускать такие числа), исключая также и 0 [+].
//[+]
//Стадия буферизации данных в кольцевом буфере с интерфейсом, соответствующим тому, который был дан в
//качестве задания в 19 модуле. [+] В этой стадии предусмотреть опустошение буфера (и соответственно, передачу
//этих данных, если они есть, дальше) с определённым интервалом во времени. [+] Значения размера буфера и этого
//интервала времени сделать настраиваемыми (как мы делали: через константы или глобальные переменные).
//[+]
//Написать источник данных для конвейера. Непосредственным источником данных должна быть консоль.
//[+]
//Также написать код потребителя данных конвейера. Данные от конвейера можно направить снова в консоль
//построчно, сопроводив их каким-нибудь поясняющим текстом, например: «Получены данные …».
//[+]
//При написании источника данных подумайте о фильтрации нечисловых данных, которые можно ввести через
//консоль. Как и где их фильтровать, решайте сами.
//=========================================================================================================

//Positive Стадия фильтрации отрицательных чисел (не пропускать отрицательные числа).
//[+]
func Positive(wg *sync.WaitGroup, done <-chan int, cIn <-chan int) <-chan int {
	cOut := make(chan int)
	wg.Add(1)
	go func() {
		for {
			select {
			case <-done:
				wg.Done()
				return

			case el := <-cIn:
				if el >= 0 {
					cOut <- el
				}
			}
		}
	}()

	return cOut
}

//Trine Стадия фильтрации чисел, не кратных 3 (не пропускать такие числа), исключая также и 0 [+].
//[+]
func Trine(wg *sync.WaitGroup, done <-chan int, cIn <-chan int) <-chan int {
	cOut := make(chan int)
	wg.Add(1)
	go func() {
		for {
			select {
			case <-done:
				wg.Done()
				return

			case el := <-cIn:
				if el%3 == 0 && el != 0 {
					cOut <- el
				}
			}
		}
	}()

	return cOut
}

//[+]
//Стадия буферизации данных в кольцевом буфере с интерфейсом, соответствующим тому, который был дан в
//качестве задания в 19 модуле. [+] В этой стадии предусмотреть опустошение буфера (и соответственно, передачу
//этих данных, если они есть, дальше) с определённым интервалом во времени. [+] Значения размера буфера и этого
//интервала времени сделать настраиваемыми (как мы делали: через константы или глобальные переменные).

//ring bufer delay
var ringDelay int = 2

//ring bufer size
var ringSize int = 5

//RingBuf make newrBuf(ringSize), work and exit
func RingBuf(wg *sync.WaitGroup, done <-chan int, cIn <-chan int) <-chan int {
	r := newrBuf(ringSize)
	cOut := make(chan int)
	wg.Add(1)
	go func() {

		for {
			select {
			case <-done:
				wg.Done()
				return

			case el := <-cIn:
				r.Pull(el)

			case <-time.After(time.Second * time.Duration(ringDelay)):
				for _, i := range r.Get() {
					cOut <- i
				}
			}
		}
	}()
	return cOut
}

// rBuf data structure
type rBuf struct {
	dataRing *ring.Ring
}

//newrBuf make rBuf
func newrBuf(size int) *rBuf {

	return &rBuf{
		ring.New(size),
	}
}

//Pull() — добавление нового элемента в буфер, в случае
//добавления в заполненный буфер стирается самый «старый» элемент
func (r *rBuf) Pull(el int) {
	r.dataRing.Value = el
	r.dataRing = r.dataRing.Next()
}

//получение всех упорядоченных элементов буфера
//и последующая очистка буфера
func (r *rBuf) Get() []int {
	resultArr := make([]int, 0)
	size := r.dataRing.Len()

	for i := 0; i < size; i++ {
		if r.dataRing.Value == nil {
			r.dataRing = r.dataRing.Next()
			continue
		}
		resultArr = append(resultArr, r.dataRing.Value.(int))
		r.dataRing.Value = nil
		r.dataRing = r.dataRing.Next()

	}
	return resultArr
}

//[+]
//Написать источник данных для конвейера. Непосредственным источником данных должна быть консоль.

// ErrNoInts - error for data check
var ErrNoInts = fmt.Errorf("There is no ints to process")

// ErrWrongCmd - error for wrong command check
var ErrWrongCmd = fmt.Errorf("Unsupported command. You can use commands: input, quit.")

//DataSrc struct
type DataSrc struct {
	data *[]int
	C    chan int
}

//NewDataSrc create DataSrc
func NewDataSrc() *DataSrc {
	d := make([]int, 0)
	return &DataSrc{&d,
		make(chan int),
	}
}

// intInput input ints to DataSrc
func (d DataSrc) intInput(i int) {
	*d.data = append(*d.data, i)
}

// Process send data from DataSrc to process.
// return error if DataSrc is empty
func (d DataSrc) Process() error {

	if len(*d.data) == 0 {
		return ErrNoInts
	}
	fmt.Println("Processing...")
	for _, i := range *d.data {
		d.C <- i
	}

	*d.data = make([]int, 0)
	return nil
}

//ScanCmd scan cmd
func (d DataSrc) ScanCmd(done chan int) {

	for {
		var cmd string
		fmt.Println("Enter command:")
		fmt.Scanln(&cmd)

		switch cmd {

		//[+]
		//При написании источника данных подумайте о фильтрации нечисловых данных,
		//которые можно ввести через консоль. Как и где их фильтровать, решайте сами.
		case "input":

			fmt.Printf("Enter numbers separated by whitespaces:\n")
			in := bufio.NewScanner(os.Stdin)
			in.Scan()
			anything := in.Text()

			someSlice := strings.Fields(anything)

			for _, some := range someSlice {

				if num, err := strconv.ParseInt(some, 10, 0); err == nil {

					d.intInput(int(num))
				}
			}

			err := d.Process()
			if err != nil {
				fmt.Println(err)
				break
			}
			time.Sleep(time.Second * time.Duration(ringDelay+1))

		case "quit":
			close(done)

			return

		default:
			fmt.Println(ErrWrongCmd)
		}
	}
}

//[+]
//Также написать код потребителя данных конвейера. Данные от конвейера можно направить снова в консоль
//построчно, сопроводив их каким-нибудь поясняющим текстом, например: «Получены данные …».

func receiver(wg *sync.WaitGroup, done chan int, c <-chan int) {
	for {
		select {
		case <-done:
			wg.Done()
			return

		case el := <-c:
			fmt.Printf("Int received: %d\n", el)
		}

	}
}

func main() {
	var wg sync.WaitGroup

	d := NewDataSrc()
	done := make(chan int)
	pipeline := RingBuf(&wg, done, Trine(&wg, done, Positive(&wg, done, d.C)))

	go d.ScanCmd(done)
	wg.Add(1)
	go receiver(&wg, done, pipeline)
	wg.Wait()
}
