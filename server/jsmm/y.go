//line parser.y:2
package jsmm

import __yyfmt__ "fmt"

//line parser.y:2
import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

//line parser.y:13
type yySymType struct {
	yys         int
	stringValue string
	expression  Expression
	list        ExpressionList
}

const tokenString = 57346
const tokenIdentifier = 57347
const tokenInteger = 57348
const tokenFloat = 57349
const tokenBoolean = 57350
const tokenNull = 57351
const tokenAnd = 57352
const tokenOr = 57353
const tokenEOF = 57354
const tokenError = 57355
const tokenEqu = 57356
const tokenNeq = 57357
const tokenLte = 57358
const tokenGte = 57359
const tokenLt = 57360
const tokenGt = 57361
const UNARY = 57362

var yyToknames = []string{
	"tokenString",
	"tokenIdentifier",
	"tokenInteger",
	"tokenFloat",
	"tokenBoolean",
	"tokenNull",
	"tokenAnd",
	"tokenOr",
	"tokenEOF",
	"tokenError",
	"tokenEqu",
	"tokenNeq",
	"tokenLte",
	"tokenGte",
	"tokenLt",
	"tokenGt",
	"'!'",
	"'+'",
	"'-'",
	"'/'",
	"'*'",
	"'%'",
	"'.'",
	"'['",
	"UNARY",
}
var yyStatenames = []string{}

const yyEofCode = 1
const yyErrCode = 2
const yyMaxDepth = 200

//line parser.y:108

type ParseError struct {
	emsg string
}

func (e *ParseError) Error() string {
	return e.emsg
}
func NewParseError(format string, args ...interface{}) *ParseError {
	return &ParseError{fmt.Sprintf(format, args...)}
}

const eof = 0

type token struct {
	ident int
	value string
}

func (t token) String() string {
	return fmt.Sprintf("<%d[%d]:%s>", t.ident, len(t.value), string(t.value))
}

type stateFn func(*lexer) stateFn

type lexer struct {
	input     string
	start     int
	pos       int
	width     int
	tokens    chan token
	state     stateFn
	ast       Expression
	lastError error
}

func (l *lexer) next() (result rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return 0
	}
	result, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return result
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
	l.width = 0
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) emit(t int) {
	// fmt.Printf("Send %d\n",t)
	l.tokens <- token{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token{
		tokenError,
		fmt.Sprintf(format, args...),
	}
	l.lastError = NewParseError(format, args...)
	return nil
}

func lexString(l *lexer) stateFn {
	for {
		switch r := l.next(); r {
		case eof:
			return l.errorf("Unterminated string")
		case '\\':
			if l.next() != '"' {
				l.backup()
			}
		case '"':
			l.backup()
			l.emit(tokenString)
			l.next()
			l.ignore()
			return lexDefault
		}
	}
	return lexDefault
}

func lexNumber(l *lexer) stateFn {
	emitToken := tokenInteger

	digits := "0123456789"

	if l.accept("0") && l.accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}

	l.acceptRun(digits)

	if l.peek() == '.' {
		l.next()
		emitToken = tokenFloat
		l.acceptRun(digits)
		if l.accept("eE") {
			l.accept("+-")
			l.acceptRun("0123456789")
		}
	}
	if unicode.IsLetter(l.peek()) {
		return l.errorf("Unexpected character in number: %c", l.peek())
	}
	l.emit(emitToken)
	return lexDefault
}

func lexIdentifier(l *lexer) stateFn {
	r := l.next()
	if !unicode.IsLetter(r) && r != '_' {
		return l.errorf("Unexpected character '%c' in identifier", r)
	}
	for {
		r := l.next()
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			l.backup()
			switch {
			case l.input[l.start:l.pos] == "true":
				l.emit(tokenBoolean)
			case l.input[l.start:l.pos] == "false":
				l.emit(tokenBoolean)
			case l.input[l.start:l.pos] == "null":
				l.emit(tokenNull)
			default:
				l.emit(tokenIdentifier)
			}
			break
		}
	}
	return lexDefault
}

func lexDefault(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case unicode.IsSpace(r):
			l.ignore()
		case unicode.IsLetter(r), r == '_':
			l.backup()
			return lexIdentifier
		case r == '"':
			l.ignore()
			return lexString
		case r <= '9' && r >= '0':
			l.backup()
			return lexNumber
		case r == '=' && l.peek() == '=':
			l.next()
			l.emit(tokenEqu)
		case r == '!' && l.peek() == '=':
			l.next()
			l.emit(tokenNeq)
		case r == '<' && l.peek() == '=':
			l.next()
			l.emit(tokenLte)
		case r == '>' && l.peek() == '=':
			l.next()
			l.emit(tokenGte)
		case r == '<':
			l.emit(tokenLt)
		case r == '>':
			l.emit(tokenGt)
		case strings.IndexRune("!+-*/%[](),:.{}", r) >= 0:
			l.emit(int(r))
		case r == '&' && l.peek() == '&':
			l.next()
			l.emit(tokenAnd)
		case r == '|' && l.peek() == '|':
			l.next()
			l.emit(tokenOr)
		case r == eof:
			l.emit(tokenEOF)
			return nil
		default:
			return l.errorf("Unexpected character '%c'", r)
		}
	}
}

func newLexer(input string) *lexer {
	l := &lexer{
		input:  input,
		tokens: make(chan token, 2),
		state:  lexDefault,
	}
	go l.run()
	return l
}

func (l *lexer) run() {
	for l.state != nil {
		l.state = l.state(l)
	}
}

func (l *lexer) Lex(lval *yySymType) int {
	//tk := l.nextToken()
	tk := <-l.tokens
	//if tk.ident==tokenError {
	//    fmt.Println("Recv Error")
	//} else {
	//    fmt.Printf("Recv %d\n",tk.ident)
	//}
	//fmt.Println(tk)
	lval.stringValue = tk.value
	return tk.ident
}
func (l *lexer) Error(e string) {
	if l.lastError == nil {
		l.lastError = NewParseError(e)
	}
}

func Parse(s string) (Expression, error) {
	l := newLexer(s)
	yyParse(l)

	if l.lastError != nil {
		return nil, l.lastError
	}
	return l.ast, nil
}

//line yacctab:1
var yyExca = []int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyNprod = 49
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 315

var yyAct = []int{

	64, 2, 85, 84, 62, 35, 36, 37, 44, 45,
	70, 69, 68, 91, 67, 74, 59, 86, 41, 74,
	82, 46, 47, 48, 49, 50, 51, 52, 53, 54,
	55, 56, 57, 58, 71, 60, 31, 32, 43, 38,
	25, 26, 29, 30, 27, 28, 42, 20, 21, 22,
	23, 24, 33, 34, 73, 39, 74, 72, 65, 66,
	20, 21, 22, 23, 24, 33, 34, 75, 33, 34,
	78, 79, 76, 77, 14, 83, 80, 22, 23, 24,
	33, 34, 13, 15, 12, 89, 90, 87, 31, 32,
	1, 0, 25, 26, 29, 30, 27, 28, 0, 20,
	21, 22, 23, 24, 33, 34, 0, 0, 61, 9,
	16, 7, 8, 10, 11, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 4, 0, 5, 0, 0,
	0, 0, 17, 0, 6, 88, 0, 0, 18, 9,
	16, 7, 8, 10, 11, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 4, 0, 5, 0, 0,
	0, 0, 17, 0, 6, 81, 0, 0, 18, 9,
	16, 7, 8, 10, 11, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 4, 0, 5, 0, 0,
	0, 0, 17, 0, 6, 63, 0, 0, 18, 9,
	16, 7, 8, 10, 11, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 4, 0, 5, 0, 0,
	0, 0, 17, 0, 6, 0, 40, 0, 18, 9,
	16, 7, 8, 10, 11, 0, 0, 3, 0, 9,
	16, 7, 8, 10, 11, 4, 0, 5, 0, 0,
	0, 0, 17, 0, 6, 4, 0, 5, 18, 0,
	0, 0, 17, 0, 6, 31, 32, 19, 18, 25,
	26, 29, 30, 27, 28, 0, 20, 21, 22, 23,
	24, 33, 34, 31, 32, 0, 0, 25, 26, 29,
	30, 27, 28, 0, 20, 21, 22, 23, 24, 33,
	34, 25, 26, 29, 30, 27, 28, 0, 20, 21,
	22, 23, 24, 33, 34,
}
var yyPact = []int{

	225, -1000, 255, -1000, 235, 235, 235, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 10, 195, 4, -1000,
	235, 235, 235, 235, 235, 235, 235, 235, 235, 235,
	235, 235, 235, 11, 235, 39, -1000, 78, 165, 27,
	-1000, 273, -20, -1000, -24, -25, 54, 54, 42, 42,
	42, 287, 287, 287, 287, 287, 287, 273, 273, 5,
	26, -1000, 24, -1000, 273, -1000, 235, -1000, 68, 235,
	235, 135, -9, -1000, 235, 273, -32, -33, 273, 273,
	-13, -1000, 105, 273, 235, 235, -1000, -17, -1000, 273,
	273, -1000,
}
var yyPgo = []int{

	0, 0, 90, 84, 83, 82, 74, 55, 4, 46,
}
var yyR1 = []int{

	0, 2, 2, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 3, 3,
	3, 4, 4, 4, 4, 4, 4, 8, 8, 5,
	5, 7, 7, 6, 6, 9, 9, 9, 9,
}
var yyR2 = []int{

	0, 2, 1, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 2, 2, 3, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 3, 4,
	1, 6, 5, 7, 6, 4, 3, 1, 3, 3,
	2, 1, 3, 3, 2, 3, 3, 5, 5,
}
var yyChk = []int{

	-1000, -2, -1, 12, 20, 22, 29, 6, 7, 4,
	8, 9, -3, -5, -6, -4, 5, 27, 33, 12,
	21, 22, 23, 24, 25, 14, 15, 18, 19, 16,
	17, 10, 11, 26, 27, -1, -1, -1, 29, -7,
	31, -1, -9, 34, 4, 5, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, 5,
	-1, 30, -8, 30, -1, 31, 32, 34, 32, 35,
	35, 29, 31, 30, 32, -1, 4, 5, -1, -1,
	-8, 30, 29, -1, 35, 35, 30, -8, 30, -1,
	-1, 30,
}
var yyDef = []int{

	0, -2, 0, 2, 0, 0, 0, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 30, 0, 0, 1,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 16, 17, 0, 0, 0,
	40, 41, 0, 44, 0, 0, 3, 4, 5, 6,
	7, 8, 9, 10, 11, 12, 13, 14, 15, 28,
	0, 18, 0, 36, 37, 39, 0, 43, 0, 0,
	0, 0, 29, 35, 0, 42, 0, 0, 45, 46,
	0, 32, 0, 38, 0, 0, 31, 0, 34, 47,
	48, 33,
}
var yyTok1 = []int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 20, 3, 3, 3, 25, 3, 3,
	29, 30, 24, 21, 32, 22, 26, 23, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 35, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 27, 3, 31, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 33, 3, 34,
}
var yyTok2 = []int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 28,
}
var yyTok3 = []int{
	0,
}

//line yaccpar:1

/*	parser for yacc output	*/

var yyDebug = 0

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

const yyFlag = -1000

func yyTokname(c int) string {
	// 4 is TOKSTART above
	if c >= 4 && c-4 < len(yyToknames) {
		if yyToknames[c-4] != "" {
			return yyToknames[c-4]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yylex1(lex yyLexer, lval *yySymType) int {
	c := 0
	char := lex.Lex(lval)
	if char <= 0 {
		c = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		c = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			c = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		c = yyTok3[i+0]
		if c == char {
			c = yyTok3[i+1]
			goto out
		}
	}

out:
	if c == 0 {
		c = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(c), uint(char))
	}
	return c
}

func yyParse(yylex yyLexer) int {
	var yyn int
	var yylval yySymType
	var yyVAL yySymType
	yyS := make([]yySymType, yyMaxDepth)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yychar := -1
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yychar), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yychar < 0 {
		yychar = yylex1(yylex, &yylval)
	}
	yyn += yychar
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yychar { /* valid shift */
		yychar = -1
		yyVAL = yylval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yychar < 0 {
			yychar = yylex1(yylex, &yylval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yychar {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error("syntax error")
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yychar))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yychar))
			}
			if yychar == yyEofCode {
				goto ret1
			}
			yychar = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		//line parser.y:40
		{
			if l, ok := yylex.(*lexer); ok {
				l.ast = yyS[yypt-1].expression
			}
			return 1
		}
	case 2:
		//line parser.y:41
		{
			return 1
		}
	case 3:
		//line parser.y:44
		{
			yyVAL.expression = &binExpr{i_add, yyS[yypt-2].expression, yyS[yypt-0].expression}
		}
	case 4:
		//line parser.y:45
		{
			yyVAL.expression = &binExpr{i_sub, yyS[yypt-2].expression, yyS[yypt-0].expression}
		}
	case 5:
		//line parser.y:46
		{
			yyVAL.expression = &binExpr{i_div, yyS[yypt-2].expression, yyS[yypt-0].expression}
		}
	case 6:
		//line parser.y:47
		{
			yyVAL.expression = &binExpr{i_mul, yyS[yypt-2].expression, yyS[yypt-0].expression}
		}
	case 7:
		//line parser.y:48
		{
			yyVAL.expression = &binExpr{i_mod, yyS[yypt-2].expression, yyS[yypt-0].expression}
		}
	case 8:
		//line parser.y:49
		{
			yyVAL.expression = &binExpr{i_equ, yyS[yypt-2].expression, yyS[yypt-0].expression}
		}
	case 9:
		//line parser.y:50
		{
			yyVAL.expression = &binExpr{i_neq, yyS[yypt-2].expression, yyS[yypt-0].expression}
		}
	case 10:
		//line parser.y:51
		{
			yyVAL.expression = &binExpr{i_lt, yyS[yypt-2].expression, yyS[yypt-0].expression}
		}
	case 11:
		//line parser.y:52
		{
			yyVAL.expression = &binExpr{i_gt, yyS[yypt-2].expression, yyS[yypt-0].expression}
		}
	case 12:
		//line parser.y:53
		{
			yyVAL.expression = &binExpr{i_lte, yyS[yypt-2].expression, yyS[yypt-0].expression}
		}
	case 13:
		//line parser.y:54
		{
			yyVAL.expression = &binExpr{i_gte, yyS[yypt-2].expression, yyS[yypt-0].expression}
		}
	case 14:
		//line parser.y:55
		{
			yyVAL.expression = &binExpr{i_and, yyS[yypt-2].expression, yyS[yypt-0].expression}
		}
	case 15:
		//line parser.y:56
		{
			yyVAL.expression = &binExpr{i_or, yyS[yypt-2].expression, yyS[yypt-0].expression}
		}
	case 16:
		//line parser.y:57
		{
			yyVAL.expression = &unaryExpr{i_not, yyS[yypt-0].expression}
		}
	case 17:
		//line parser.y:58
		{
			yyVAL.expression = &unaryExpr{i_neg, yyS[yypt-0].expression}
		}
	case 18:
		//line parser.y:59
		{
			yyVAL.expression = yyS[yypt-1].expression
		}
	case 19:
		//line parser.y:60
		{
			yyVAL.expression = &literalNumberExpr{yyS[yypt-0].stringValue}
		}
	case 20:
		//line parser.y:61
		{
			yyVAL.expression = &literalNumberExpr{yyS[yypt-0].stringValue}
		}
	case 21:
		//line parser.y:62
		{
			yyVAL.expression = &literalStringExpr{yyS[yypt-0].stringValue}
		}
	case 22:
		//line parser.y:63
		{
			yyVAL.expression = &literalBooleanExpr{yyS[yypt-0].stringValue}
		}
	case 23:
		//line parser.y:64
		{
			yyVAL.expression = &literalNullExpr{}
		}
	case 24:
		//line parser.y:65
		{
			yyVAL.expression = yyS[yypt-0].expression
		}
	case 25:
		//line parser.y:66
		{
			yyVAL.expression = yyS[yypt-0].expression
		}
	case 26:
		//line parser.y:67
		{
			yyVAL.expression = yyS[yypt-0].expression
		}
	case 27:
		//line parser.y:68
		{
			yyVAL.expression = yyS[yypt-0].expression
		}
	case 28:
		//line parser.y:72
		{
			yyVAL.expression = &attributeSelectionExpr{yyS[yypt-2].expression, &literalStringExpr{yyS[yypt-0].stringValue}}
		}
	case 29:
		//line parser.y:73
		{
			yyVAL.expression = &attributeSelectionExpr{yyS[yypt-3].expression, yyS[yypt-1].expression}
		}
	case 30:
		//line parser.y:74
		{
			yyVAL.expression = &attributeSelectionExpr{&getGlobalObjectExpr{}, &literalStringExpr{yyS[yypt-0].stringValue}}
		}
	case 31:
		//line parser.y:77
		{
			yyVAL.expression = &functionCallExpr{yyS[yypt-5].expression, &literalStringExpr{yyS[yypt-3].stringValue}, yyS[yypt-1].list}
		}
	case 32:
		//line parser.y:78
		{
			yyVAL.expression = &functionCallExpr{yyS[yypt-4].expression, &literalStringExpr{yyS[yypt-2].stringValue}, NewExpressionList()}
		}
	case 33:
		//line parser.y:79
		{
			yyVAL.expression = &functionCallExpr{yyS[yypt-6].expression, yyS[yypt-4].expression, yyS[yypt-1].list}
		}
	case 34:
		//line parser.y:80
		{
			yyVAL.expression = &functionCallExpr{yyS[yypt-5].expression, yyS[yypt-3].expression, NewExpressionList()}
		}
	case 35:
		//line parser.y:81
		{
			yyVAL.expression = &functionCallExpr{&getGlobalObjectExpr{}, &literalStringExpr{yyS[yypt-3].stringValue}, yyS[yypt-1].list}
		}
	case 36:
		//line parser.y:82
		{
			yyVAL.expression = &functionCallExpr{&getGlobalObjectExpr{}, &literalStringExpr{yyS[yypt-2].stringValue}, NewExpressionList()}
		}
	case 37:
		//line parser.y:85
		{
			yyVAL.list = NewExpressionList().Append(yyS[yypt-0].expression)
		}
	case 38:
		//line parser.y:86
		{
			yyVAL.list = yyS[yypt-2].list.Append(yyS[yypt-0].expression)
		}
	case 39:
		//line parser.y:89
		{
			yyVAL.expression = &arrayDefExpr{yyS[yypt-1].list}
		}
	case 40:
		//line parser.y:90
		{
			yyVAL.expression = &arrayDefExpr{nil}
		}
	case 41:
		//line parser.y:93
		{
			yyVAL.list = NewExpressionList().Append(&unaryExpr{i_array_append, yyS[yypt-0].expression})
		}
	case 42:
		//line parser.y:94
		{
			yyVAL.list = yyS[yypt-2].list.Append(&unaryExpr{i_array_append, yyS[yypt-0].expression})
		}
	case 43:
		//line parser.y:97
		{
			yyVAL.expression = &objectDefExpr{yyS[yypt-1].list}
		}
	case 44:
		//line parser.y:98
		{
			yyVAL.expression = &objectDefExpr{nil}
		}
	case 45:
		//line parser.y:101
		{
			yyVAL.list = NewExpressionList().Append(&binExpr{i_set_index, &literalStringExpr{yyS[yypt-2].stringValue}, yyS[yypt-0].expression})
		}
	case 46:
		//line parser.y:102
		{
			yyVAL.list = NewExpressionList().Append(&binExpr{i_set_index, &literalStringExpr{yyS[yypt-2].stringValue}, yyS[yypt-0].expression})
		}
	case 47:
		//line parser.y:103
		{
			yyVAL.list = yyS[yypt-4].list.Append(&binExpr{i_set_index, &literalStringExpr{yyS[yypt-2].stringValue}, yyS[yypt-0].expression})
		}
	case 48:
		//line parser.y:104
		{
			yyVAL.list = yyS[yypt-4].list.Append(&binExpr{i_set_index, &literalStringExpr{yyS[yypt-2].stringValue}, yyS[yypt-0].expression})
		}
	}
	goto yystack /* stack new state and value */
}
