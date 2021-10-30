package controller

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func StartPage(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	text := "Приветствую тебя на стартовой странице этого сайта!"
	//возвращаем простой текст
	fmt.Fprint(rw, text)
}
