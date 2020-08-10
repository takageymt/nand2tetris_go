// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/04/Fill.asm

// Runs an infinite loop that listens to the keyboard input.
// When a key is pressed (any key), the program blackens the screen,
// i.e. writes "black" in every pixel;
// the screen should remain fully black as long as the key is pressed. 
// When no key is pressed, the program clears the screen, i.e. writes
// "white" in every pixel;
// the screen should remain fully clear as long as no key is pressed.

// Put your code here.
(LOOP)
	@SCREEN
	D=A
	@now
	M=D
	@row
	M=0
	@bits
	M=-1 // 1111 1111 1111 1111
	@KBD
	D=M
	@FILL
	D;JNE
	@bits
	M=0 // 0000 0000 0000 0000
(FILL)
	@row
	D=M
	@256
	D=D-A
	@DONE
	D;JGE

	@col
	M=0
(COL_FILL)
	@col
	D=M
	@32
	D=D-A
	@COL_DONE
	D;JGE
	@bits
	D=M
	@now
	A=M
	M=D

	@now
	M=M+1
	@col
	M=M+1
	@COL_FILL
	0;JMP
(COL_DONE)

	@row
	M=M+1
	@FILL
	0;JMP
(DONE)
	@LOOP
	0;JMP
