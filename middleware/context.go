package middleware

import(
	"net/http"
	"context"
	"time"
)


/*	Input: 	function of type http.HandlerFunc 
*	Output: function of type http.HandlerFunc with modified context
*
*	Note: Remember, HandlerFunc is an adapter for functions func(ResponseWriter, *Request)
*
*/

func ModHandlerContext(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		
		// create a context
		ctx := context.Background()
		
		// set timeout
		ctx, cancel := context.WithTimeout(ctx, 10000 * time.Millisecond)	
		//make sure cancel is called, as all requests must end at some point regardless of outcome
		defer cancel()
		
		// TODO: perform any other modifications to context as needed

		// assign the new context to the http request
		r = r.WithContext(ctx)

		//run the ServerHTTP function 
		h.ServeHTTP(w, r) 
	})
}