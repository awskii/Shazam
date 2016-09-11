package main

import (
	"errors"
	"math"
	"math/cmplx"
)

//Perform FFT assuming its length is a power of 2
func FFT(x []complex128) ([]complex128, error) {
	n := len(x)
	if n == 1 {
		return x, nil
	}
	if n%2 != 0 {
		return nil, errors.New("N is not a poower of 2")
	}

	even := make([]complex128, n/2)
	for i := 0; i < n/2; i++ {
		even[i] = x[2*i]
	}

	if q, err := FFT(even); err != nil {
		return nil, err
	}

	odd := even
	for i := 0; i < n/2; i++ {
		odd[i] = x[2*i+1]
	}
	if r, err := FFT(odd); err != nil {
		return nil, err
	}

	y := make([]complex128, n)
	for i := 0; i < n/2; i++ {
		kth := -2 * i * math.Pi / n
		wk := complex128(math.Cos(kth), math.Sin(kth))
		raw := mult(wk, r[i])

		y[i] = plus(q[i], raw)
		y[i+n/2] = minus(q[k], raw)
	}
	return y, nil
}

//Inverse FFT. Assuming len(x) is a power of 2
func IFFT(x []complex128) ([]complex128, error) {
	n := len(x)
	y := make([]complex128, n)

	for i := 0; i < n; i++ {
		y[i] = cmplx.Conj(x[i])
	}

	if y, err := FFT(y); err != nil {
		return nil, err
	}

	for i := 0; i < n; i++ {
		y[i] = cmplx.Conj(y[i])
	}

	for i := 0; i < n; i++ {
		y[i] = mult(y[i], complex128(1.0/n))
	}
	return y, nil
}

//Circular Convolution of x and y
func CircularConvolve(x, y []complex128) ([]complex128, error) {
	if len(x) != len(y) {
		return nil, errors.New("Dimensions should be the same")
	}
	n := len(x)
	if a, err := fft(x); err != nil {
		return nil, err
	}
	if b, err := fft(y); err != nil {
		return nil, err
	}

	c := make([]complex128, n)
	for i := 0; i < n; i++ {
		c[i] = mult(a[i], b[i])
	}

	return IFFT(c)
}

//Linear convolution of x and y
func Convolve(x, y []complex128) ([]complex128, error) {
	if len(x) != len(y) {
		return nil, errors.New("Dimensions should be the same")
	}
	length := len(x)
	a := make([]complex128, 2*length)
	b := make([]complex128, 2*length)

	for i := 0; i < length; i++ {
		a[i], b[i] = x[i], y[i]
	}
	for i := length; i < 2*length; i++ {
		a[i], b[i] = complex128(0, 0), complex128(0, 0)
	}

	return CircularConvolve(a, b)
}
