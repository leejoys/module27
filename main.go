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

//Задача:
//измените код так, чтобы каждое действие в программе пайплайна (включая каждую его стадию)
//добавлялось в логи. В качестве потока, в который нужно направлять эти логи, выберите консоль
//=========================================================================================================

//logger receive log messages and write them into logStream
func logger(log <-chan string) {

	//define where to write the log
	var logStream *os.File = os.Stdout

	for {
		in := <-log
		switch in {
		case "end of log":
			fmt.Fprintln(logStream, in)
			return

		default:
			fmt.Fprintln(logStream, in)
		}
	}
}

//Positive Стадия фильтрации отрицательных чисел (не пропускает отрицательные числа).
func Positive(log chan<- string, wg *sync.WaitGroup, done <-chan int, cIn <-chan int) <-chan int {
	cOut := make(chan int)
	wg.Add(1)
	go func() {
		for {
			select {
			case <-done:
				log <- "stop Positive"
				wg.Done()
				return

			case el := <-cIn:
				if el >= 0 {
					log <- fmt.Sprintf("Positive stage passed by %d", el)
					cOut <- el
				} else {
					log <- fmt.Sprintf("Positive stage failed by %d", el)
				}
			}
		}
	}()

	return cOut
}

//Trine Стадия фильтрации чисел, не кратных 3 (не пропускает такие числа), исключая также и 0.
func Trine(log chan<- string, wg *sync.WaitGroup, done <-chan int, cIn <-chan int) <-chan int {
	cOut := make(chan int)
	wg.Add(1)
	go func() {
		for {
			select {
			case <-done:
				log <- "stop Trine"
				wg.Done()
				return

			case el := <-cIn:
				if el%3 == 0 && el != 0 {
					log <- fmt.Sprintf("Trine stage passed by %d", el)
					cOut <- el
				} else {
					log <- fmt.Sprintf("Trine stage failed by %d", el)
				}
			}
		}
	}()

	return cOut
}

//Стадия буферизации данных в кольцевом буфере с интерфейсом, соответствующим тому, который был дан в
//качестве задания в 19 модуле. В этой стадии предусмотрено опустошение буфера (и соответственно, передачу
//этих данных, если они есть, дальше) с определённым интервалом во времени. Значения размера буфера и этого
//интервала времени настраивается через глобальные переменные.

//ring bufer delay
var ringDelay int = 2

//ring bufer size
var ringSize int = 5

//RingBuf make newrBuf(ringSize), work and exit
func RingBuf(log chan<- string, wg *sync.WaitGroup, done <-chan int, cIn <-chan int) <-chan int {
	r := newrBuf(ringSize)
	cOut := make(chan int)
	wg.Add(1)
	go func() {

		for {
			select {
			case <-done:
				log <- "stop RingBuf"
				wg.Done()
				return

			case el := <-cIn:
				r.Pull(el)
				log <- fmt.Sprintf("%d added to RingBuf", el)

			case <-time.After(time.Second * time.Duration(ringDelay)):
				for _, i := range r.Get() {
					log <- fmt.Sprintf("RingBuf release: %d", i)
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

//Источник данных для конвейера. Непосредственный источник данных - консоль.

// ErrNoInts - error for data check
var ErrNoInts = fmt.Errorf("there is no ints to process")

// ErrWrongCmd - error for wrong command check
var ErrWrongCmd = fmt.Errorf("unsupported command, you can use commands: input, quit")

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

// process send data from DataSrc to process.
// return error if DataSrc is empty
func (d DataSrc) process() error {

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
func (d DataSrc) ScanCmd(log chan<- string, done chan int) {

	for {
		var cmd string
		fmt.Println("Enter command:")
		fmt.Scanln(&cmd)
		log <- fmt.Sprintf("command received: %s", cmd)

		switch cmd {

		//Фильтрация нечисловых данных, которые можно ввести через консоль.
		case "input":
			log <- "INPUT NUMBERS mode"
			fmt.Printf("Enter numbers separated by whitespaces:\n")
			in := bufio.NewScanner(os.Stdin)
			in.Scan()
			anything := in.Text()
			log <- fmt.Sprintf("string received: %s", anything)

			someSlice := strings.Fields(anything)

			for _, some := range someSlice {

				if num, err := strconv.ParseInt(some, 10, 0); err == nil {

					log <- fmt.Sprintf("number sent to processing: %d", (int(num)))
					d.intInput(int(num))
				}
			}

			err := d.process()
			if err != nil {
				fmt.Println(err)
				log <- fmt.Sprintf("error: %s", err)
				break
			}
			time.Sleep(time.Second * time.Duration(ringDelay+1))

		case "quit":
			log <- "sending DONE signal and stop ScanCmd"
			close(done)
			return

		default:
			log <- fmt.Sprintf("error: %s", ErrWrongCmd)
			fmt.Println(ErrWrongCmd)
		}
	}
}

//Код потребителя данных конвейера. Данные от конвейера направляются снова в консоль
//построчно, с поясняющим текстом: «Получены данные …».

func receiver(log chan<- string, wg *sync.WaitGroup, done chan int, c <-chan int) {
	for {
		select {
		case <-done:
			log <- "stop receiver"
			wg.Done()
			return

		case el := <-c:
			log <- fmt.Sprintf("number received by receiver: %d", el)
			fmt.Printf("Int received: %d\n", el)
		}
	}
}

func main() {
	var wg sync.WaitGroup

	//logger start
	logChan := make(chan string)
	go logger(logChan)
	logChan <- "program started"

	d := NewDataSrc()
	done := make(chan int)
	pipeline := RingBuf(logChan, &wg, done, Trine(logChan, &wg, done, Positive(logChan, &wg, done, d.C)))

	go d.ScanCmd(logChan, done)
	wg.Add(1)
	go receiver(logChan, &wg, done, pipeline)
	wg.Wait()

	//logger stop
	logChan <- "end of log"
	time.Sleep(time.Second / 100)
}
