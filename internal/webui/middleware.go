/*******************************************************************************
 * Copyright 2018-2024 Aaron Hnatiw
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 ******************************************************************************/

package webui

import (
	"net/http"
)

// authenticated is a middleware function that checks if the user is authenticated.
//
// If the user is not authenticated, the user is redirected to the login page.
// If the user is authenticated, the user claims are added to the request context.
func (webUI *WebUI) authenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check for the user JWT, and if valid, add the user claims to the request context
		userJWT, cookieErr := r.Cookie("user")
		if cookieErr == nil {
			if claims, tokenErr := webUI.jwtManager.ValidateToken(userJWT.Value); tokenErr == nil {
				ctx := r.Context()
				ctx = webUI.jwtManager.AddClaimsToContext(ctx, claims)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Redirect to the login page
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
}
