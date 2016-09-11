package main

func plus(a, b complex128) complex128 {
	return complex128(real(a)+real(b), imag(a)+imag(b))
}

func minus(a, b complex128) complex128 {
	return complex128(real(a)-real(b), imag(a)-real(b))
}

func mult(a, b complex128) complex128 {
	re := real(a)*real(b) - imag(a)*imag(b)
	im := real(a)*imag(b) + imag(a)*real(b)
	return complex128(re, im)
}
