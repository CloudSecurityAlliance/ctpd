%{
package jsmm

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

%}

%union{
    stringValue    string
    expression     Expression
    list           ExpressionList
}

%token tokenString tokenIdentifier tokenInteger tokenFloat tokenBoolean 
/* %token tokenEqu tokenNeq tokenGt tokenLt tokenGte tokenLte */
%token tokenAnd tokenOr tokenEOF tokenError

%type <expression> expr final attribute call arraydef objectdef
%type <list> alist clist olist
%type <stringValue> tokenString tokenIdentifier tokenInteger tokenFloat tokenBoolean
/*%type <list> listDecl*/
/* %type <call> fc fcall */

%right tokenOr tokenAnd
%right tokenEqu tokenNeq tokenLte tokenGte tokenLt tokenGt
%right '!'
%left '+' '-'
%left '/' '*' '%'
%left '.'
%left '['
%left UNARY

%%

final: expr tokenEOF { if l,ok:=yylex.(*lexer); ok { l.ast=$1 }; return 1 }
     | tokenEOF      { return 1 }
     ;

expr: expr '+' expr             { $$=&binExpr{i_add,$1,$3} }
    | expr '-' expr             { $$=&binExpr{i_sub,$1,$3} }
    | expr '/' expr             { $$=&binExpr{i_div,$1,$3} }
    | expr '*' expr             { $$=&binExpr{i_mul,$1,$3} }
    | expr '%' expr             { $$=&binExpr{i_mod,$1,$3} }
    | expr tokenEqu expr        { $$=&binExpr{i_equ,$1,$3} }
    | expr tokenNeq expr        { $$=&binExpr{i_neq,$1,$3} }
    | expr tokenLt expr         { $$=&binExpr{i_lt,$1,$3} }
    | expr tokenGt expr         { $$=&binExpr{i_gt,$1,$3} }
    | expr tokenLte expr        { $$=&binExpr{i_lte,$1,$3} }
    | expr tokenGte expr        { $$=&binExpr{i_gte,$1,$3} }
    | expr tokenAnd expr        { $$=&binExpr{i_and,$1,$3} }
    | expr tokenOr expr         { $$=&binExpr{i_or,$1,$3} }
    | '!' expr                  { $$=&unaryExpr{i_not,$2} }
    | '-' expr  %prec UNARY     { $$=&unaryExpr{i_neg,$2} }
    | '(' expr ')'              { $$=$2 }
    | tokenInteger              { $$=&literalNumberExpr{$1} }
    | tokenFloat                { $$=&literalNumberExpr{$1} }
    | tokenString               { $$=&literalStringExpr{$1} }
    | tokenBoolean              { $$=&literalBooleanExpr{$1} }
    | attribute                 { $$=$1 }
    | arraydef                  { $$=$1 }
    | objectdef                 { $$=$1 }
    | call                      { $$=$1 }
    ;


attribute: expr '.' tokenIdentifier    { $$=&attributeSelectionExpr{$1,&literalStringExpr{$3}} }
         | expr '[' expr ']'           { $$=&attributeSelectionExpr{$1,$3} }
         | tokenIdentifier             { $$=&attributeSelectionExpr{&getGlobalObjectExpr{},&literalStringExpr{$1}} }
         ;

call:   expr '.' tokenIdentifier '(' clist ')'   { $$=&functionCallExpr{$1,&literalStringExpr{$3},$5} }
    |   expr '.' tokenIdentifier '(' ')'         { $$=&functionCallExpr{$1,&literalStringExpr{$3},NewExpressionList()} }
    |   expr '[' expr ']' '(' clist ')'          { $$=&functionCallExpr{$1,$3,$6} }
    |   expr '[' expr ']' '(' ')'                { $$=&functionCallExpr{$1,$3,NewExpressionList()} }
    |   tokenIdentifier '(' clist ')'            { $$=&functionCallExpr{&getGlobalObjectExpr{},&literalStringExpr{$1},$3} }
    |   tokenIdentifier '(' ')'                  { $$=&functionCallExpr{&getGlobalObjectExpr{},&literalStringExpr{$1},NewExpressionList()} }
    ;

clist: expr               { $$=NewExpressionList().Append($1) }
     | clist ',' expr     { $$=$1.Append($3) }
     ;

arraydef:   '[' alist ']'       { $$=&arrayDefExpr{$2} }
        |   '[' ']'             { $$=&arrayDefExpr{nil} }
        ;

alist: expr               { $$=NewExpressionList().Append(&unaryExpr{i_array_append,$1}) }
     | alist ',' expr     { $$=$1.Append(&unaryExpr{i_array_append,$3}) }
     ;

objectdef:  '{' olist '}'       { $$=&objectDefExpr{$2} }
         |  '{' '}'             { $$=&objectDefExpr{nil} }
         ;

olist: tokenString ':' expr                 { $$=NewExpressionList().Append(&binExpr{i_set_index,&literalStringExpr{$1},$3}) }
     | tokenIdentifier ':' expr             { $$=NewExpressionList().Append(&binExpr{i_set_index,&literalStringExpr{$1},$3}) }
     | olist ',' tokenString ':' expr       { $$=$1.Append(&binExpr{i_set_index,&literalStringExpr{$3},$5}) }
     | olist ',' tokenIdentifier ':' expr   { $$=$1.Append(&binExpr{i_set_index,&literalStringExpr{$3},$5}) }
     ;


%%

type ParseError struct {
    emsg string
}
func (e *ParseError)Error() string {
    return e.emsg
}
func NewParseError(format string, args ...interface{}) *ParseError {
    return &ParseError{fmt.Sprintf(format,args...)}
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
	input       string
	start       int
	pos         int
	width       int
	tokens      chan token
	state       stateFn
    ast         Expression
    lastError   error
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

func lexIdentifier(l* lexer) stateFn {
    if !unicode.IsLetter(l.next()) {
        return l.errorf("Unexpected character in symbol")
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
        case unicode.IsLetter(r):
            l.backup()
            return lexIdentifier
		case r == '"':
			l.ignore()
			return lexString
		case r <= '9' && r >= '0':
			l.backup()
			return lexNumber
        case r== '=' && l.peek()=='=':
            l.next()
            l.emit(tokenEqu)
        case r== '!' && l.peek()=='=':
            l.next()
            l.emit(tokenNeq)
        case r== '<' && l.peek()=='=':
            l.next()
            l.emit(tokenLte)
        case r== '>' && l.peek()=='=':
            l.next()
            l.emit(tokenGte)
        case r== '<':
            l.emit(tokenLt)
        case r== '>':
            l.emit(tokenGt)
        case strings.IndexRune("!+-*/%[](),:.{}",r)>=0:
            l.emit(int(r))
        case r== '&' && l.peek()=='&':
            l.next()
            l.emit(tokenAnd)
        case r== '|' && l.peek()=='|':
            l.next()
            l.emit(tokenOr)
		case r == eof:
			l.emit(tokenEOF)
            return nil 
		default:
            return l.errorf("Unexpected character '%c'",r)
        }
	}
}

func newLexer(input string) *lexer {
    l := &lexer{
        input: input, 
        tokens: make(chan token, 2), 
        state: lexDefault,
    }
    go l.run()
    return l
}

func (l *lexer) run() {
    for l.state!=nil {
        l.state = l.state(l)
    }
}

func (l *lexer)Lex(lval *yySymType) int {
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
func (l *lexer)Error(e string) {
    if l.lastError==nil {
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
