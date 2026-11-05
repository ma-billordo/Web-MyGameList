package handlers

import db "MyGameList.com/db/sqlc"

type Server struct {
	queries *db.Queries
}

func NewServer(q *db.Queries) *Server {
	return &Server{queries: q}
}
