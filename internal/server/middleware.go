package server

import (
	"fmt"
	"net/http"
)

func (s *QueueServer) LeaderRedirectMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.RaftNode.IsLeader() {
			http.Error(w, fmt.Sprintf("Not the leader. Current leader: %s", s.RaftNode.GetLeader()), http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}
