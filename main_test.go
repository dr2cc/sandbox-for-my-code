package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_urlServer_GetHandler(t *testing.T) {
	tt := []struct {
		name       string
		method     string
		input      *UrlStorage
		want       string
		statusCode int
	}{
		{
			name:   "должен работать",
			method: http.MethodGet,
			input: &UrlStorage{
				Data: map[string]string{"6ba7b811": "https://practicum.yandex.ru/"},
			},
			want:       "https://practicum.yandex.ru/",
			statusCode: http.StatusTemporaryRedirect,
		},
		// {
		// 	name:       "with bad method",
		// 	method:     http.MethodPost,
		// 	input:      &Pizzas{},
		// 	want:       "Method not allowed",
		// 	statusCode: http.StatusMethodNotAllowed,
		// }, //http.StatusBadRequest,
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			responseRecorder := httptest.NewRecorder()
			//Так рекомендуют
			request := httptest.NewRequest(tc.method, "/6ba7b811", nil)

			// //Вроде если запрос от клиента, то можно использовать пакет http
			// //Но не работает. Оставлю, вдруг что-то подскажет.
			// request, _ := http.NewRequest(tc.method, "/6ba7b811", nil)

			//Вызываем метод GetHandler структуры UrlStorage
			//Этот метод делает запись в responseRecorder
			tc.input.GetHandler(responseRecorder, request)

			// По заданию на конечную точку с методом GET в инкременте 1:
			// В случае успешной обработки запроса сервер возвращает статус с кодом 307
			// и URL (переоеданный ранее) в заголовке "Location"
			if responseRecorder.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}

			if strings.TrimSpace(responseRecorder.Header()["Location"][0]) != tc.want {
				t.Errorf("Want '%s', got '%s'", tc.want, responseRecorder.Body)
			}
		})
	}
}
