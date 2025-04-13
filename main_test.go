package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_urlServer_GetHandler(t *testing.T) {
	// type args struct {
	// 	w   http.ResponseWriter
	// 	req *http.Request
	// }
	// tests := []struct {
	// 	name string
	// 	ts   *urlServer
	// 	args args
	// }{
	// 	// TODO: Add test cases.
	// }
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		tt.ts.GetHandler(tt.args.w, tt.args.req)
	// 	})
	// }

	tt := []struct {
		name   string
		method string
		// //вот input нужно заменить на мой.. 12.04.2025
		// input      *Pizzas
		input      *UrlStorage
		want       string
		statusCode int
	}{
		// {
		// 	name:       "without pizzas",
		// 	method:     http.MethodGet,
		// 	input:      &Pizzas{},
		// 	want:       "Error: No pizzas found",
		// 	statusCode: http.StatusNotFound,
		// },
		// {
		// 	name:   "with pizzas",
		// 	method: http.MethodGet,
		// 	input: &Pizzas{
		// 		Pizza{
		// 			ID:    1,
		// 			Name:  "Foo",
		// 			Price: 10,
		// 		},
		// 	},
		// 	want:       `[{"id":1,"name":"Foo","price":10}]`,
		// 	statusCode: http.StatusOK,
		// },
		// {
		// 	name:       "with bad method",
		// 	method:     http.MethodPost,
		// 	input:      &Pizzas{},
		// 	want:       "Method not allowed",
		// 	statusCode: http.StatusMethodNotAllowed,
		// },
		{
			name:   "drk 01",
			method: http.MethodGet,
			input: &UrlStorage{
				Data: map[string]string{"/6ba7b811": "https:///practicum.yandex.ru/"},
			},
			want:       "https:///practicum.yandex.ru/",
			statusCode: http.StatusTemporaryRedirect,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, "/", nil)
			responseRecorder := httptest.NewRecorder()

			//mux.Handle("/pizzas", pizzasHandler{&pizzas})
			//Вызываем метод ServeHTTP структуры pizzasHandler
			//Этот метод делает запись в responseRecorder
			//pizzasHandler{tc.input}.ServeHTTP(responseRecorder, request)

			//13.04.2025 Я никуда не передаю tc.input (мой объект хранилища)
			//Переделать.
			NewStorageStruct().GetHandler(responseRecorder, request)

			if responseRecorder.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			if strings.TrimSpace(responseRecorder.Body.String()) != tc.want {
				t.Errorf("Want '%s', got '%s'", tc.want, responseRecorder.Body)
			}
		})
	}
}
