package middleware

import (
	"PBGW/internal"
	"net"
	"net/http"
)

// Maintenance Maintenance는 유지보수 모드 미들웨어입니다.
// 서버의 서비스 상태 설정에 따라 요청을 거부 또는 허용할 수 있는 기능을 추가 할 수 있습니다.

func Maintenance(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 서버 Config 획득
		configure := internal.NewServerConfigure()

		// 요청자 IP 주소 획득
		requestIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Invalid IP address", http.StatusBadRequest)
			return
		}

		requestAllowed := false

		// 개발환경의 서버로 판단 될 경우 지정된 IP 또는 지정된 범위의 IP 주소를 허용합니다.
		if configure.Development == true {
			if IsWithinIPFilter(requestIP) == true {
				requestAllowed = true
			}
		} else {
			// 운영 환경이지만 특정 서버셋 세팅에서 별도의 관리를 하는 경우가 있을 경우 추가 구현을 진행한다.
			// 서버의 운영 상태를 관리자 툴 등을 통해서 동적 관리를 하는 경우가 있을 경우 추가 구현을 진행한다.

			// 운영환경의 서버로 판단 될 경우 블랙리스트 IP 주소를 제외한 모든 IP 주소를 허용합니다.
			if IsBlacklistedIP(requestIP) == false {
				requestAllowed = true
			}
		}

		if requestAllowed == false {
			http.Error(w, "Server is under maintenance", http.StatusServiceUnavailable)
			return
		}

		next.ServeHTTP(w, r)
	})
}
