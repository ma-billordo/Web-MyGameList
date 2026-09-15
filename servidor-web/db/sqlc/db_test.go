package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func abrirDB() (*sql.DB, error) {
	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		return nil, fmt.Errorf("La variable de entorno DATABASE_URL no está seteada\n")
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	return db, nil

}

func TestGameList_CRUD(t *testing.T) {
	conn, err := abrirDB()
	if err != nil {
		t.Fatalf("Error al iniciar la BD: (%v)", err)
	}
	defer conn.Close()

	queries := New(conn)
	ctx := context.Background()

	var userID, gameID int32

	// ---------- USERS ----------
	success := t.Run("CreateUser", func(t *testing.T) {
		userID, err = queries.CreateUser(ctx, CreateUserParams{
			Name:     "max",
			Email:    "max@gmail.com",
			Password: "1234",
		})
		if err != nil {
			t.Fatalf("Error al crear usuario: %v\n", err)
		}
	})
	if !success {
		t.Fatalf("No se pudo crear el usuario. Abortando el resto del test\n")
	}

	t.Run("GetUser", func(t *testing.T) {
		u, err := queries.GetUser(ctx, userID)
		if err != nil {
			t.Fatalf("Error al obtener usuario: %v\n", err)
		}
		fmt.Printf("ID del usuario: %d\n", u.IDUser)
		if u.Name != "max" {
			t.Errorf("Name esperado %q, obtenido %q\n", "max", u.Name)
		} else {
			fmt.Printf("Nombre del usuario: %s\n", u.Name)
		}
		if u.Email != "max@gmail.com" {
			t.Errorf("Email esperado %q, obtenido %q\n", "max@gmail.com", u.Email)
		} else {
			fmt.Printf("Email del usuario: %s\n", u.Email)
		}
	})

	t.Run("UpdateUser", func(t *testing.T) {
		err := queries.UpdateUser(ctx, UpdateUserParams{
			IDUser: userID,
			Name:   "valen",
			Email:  "valen@gmail.com",
		})
		if err != nil {
			t.Fatalf("Error al actualizar usuario: %v\n", err)
		}
		// Leer actualizado
		u, err := queries.GetUser(ctx, userID)
		if err != nil {
			t.Fatalf("Error al releer usuario: %v\n", err)
		}
		fmt.Printf("ID del usuario: %d\n", u.IDUser)
		if u.Name != "valen" {
			t.Errorf("Name esperado %q, obtenido %q\n", "valen", u.Name)
		} else {
			fmt.Printf("Nombre del usuario: %s\n", u.Name)
		}
		if u.Email != "valen@gmail.com" {
			t.Errorf("Email esperado %q, obtenido %q\n", "valen@gmail.com", u.Email)
		} else {
			fmt.Printf("Email del usuario: %s\n", u.Email)
		}
	})

	t.Run("ListUsers", func(t *testing.T) {
		users, err := queries.ListUsers(ctx)
		if err != nil {
			t.Fatalf("Error al listar usuarios: %v\n", err)
		}
		found := false
		fmt.Println("-----LISTADO-----")
		for _, u := range users {
			fmt.Printf("ID: %v  Nombre: %s  Email: %s\n", u.IDUser, u.Name, u.Email)
			fmt.Println("--------")
			if u.IDUser == userID {
				found = true
			}
		}
		if !found {
			t.Errorf("El usuario creado (ID %d) no aparece en ListUsers\n", userID)
		}
	})

	// ---------- LIBRARY (Games) ----------
	success = t.Run("CreateGame", func(t *testing.T) {
		gameID, err = queries.CreateGame(ctx, CreateGameParams{
			Title:       "Counter Strike 2",
			Description: "FPS táctico, Multijugador",
		})
		if err != nil {
			t.Fatalf("Error al crear juego: %v\n", err)
		}
	})
	if !success {
		t.Fatalf("No se pudo crear el juego. Abortando el resto del test\n")
	}

	t.Run("GetGame", func(t *testing.T) {
		g, err := queries.GetGame(ctx, gameID)
		if err != nil {
			t.Fatalf("Error al obtener juego: %v\n", err)
		}
		fmt.Printf("ID del juego: %d\n", g.IDGame)
		if g.Title != "Counter Strike 2" {
			t.Errorf("Title esperado %q, obtenido %q\n", "Counter Strike 2", g.Title)
		} else {
			fmt.Printf("Titulo del juego: %s\n", g.Title)
		}
		fmt.Printf("Descripción del juego: %s\n", g.Description)
		if !g.AvgScore.Valid || g.AvgScore.String != "0.00" {
			t.Errorf("AvgScore esperado %q, ontenido %q\n", "0.00", g.AvgScore.String)
		} else {
			fmt.Printf("Puntuación media del juego: %s\n", g.AvgScore.String)
		}

	})

	t.Run("UpdateGame", func(t *testing.T) {
		err := queries.UpdateGame(ctx, UpdateGameParams{
			IDGame:      gameID,
			Title:       "PAYDAY 2",
			Description: "FPS, Cooperativo Online",
		})
		if err != nil {
			t.Fatalf("Error al actualizar juego: %v\n", err)
		}
		// Leer actualizado
		g, err := queries.GetGame(ctx, gameID)
		if err != nil {
			t.Fatalf("Error al releer juego: %v\n", err)
		}
		fmt.Printf("ID del juego: %d\n", g.IDGame)
		if g.Title != "PAYDAY 2" {
			t.Errorf("Title esperado %q, obtenido %q", "PAYDAY 2", g.Title)
		} else {
			fmt.Printf("Titulo del juego: %s\n", g.Title)
		}
		fmt.Printf("Descripción del juego: %s\n", g.Description)
		fmt.Printf("Puntuación media del juego: %s\n", g.AvgScore.String)
	})

	t.Run("ListGames", func(t *testing.T) {
		games, err := queries.ListGames(ctx)
		if err != nil {
			t.Fatalf("Error al listar juegos: %v\n", err)
		}
		found := false
		fmt.Println("-----LISTADO-----")
		for _, g := range games {
			fmt.Printf("ID: %v  Titulo: %s\nDescripcion: %s\nPuntuación promedio: %s\n", g.IDGame, g.Title, g.Description, g.AvgScore.String)
			fmt.Println("--------")
			if g.IDGame == gameID {
				found = true
			}
		}
		if !found {
			t.Errorf("El juego creado (ID %d) no aparece en ListGames\n", gameID)
		}
	})

	// ---------- CATALOGUE (Record) ----------
	success = t.Run("CreateRecord", func(t *testing.T) {
		err := queries.CreateRecord(ctx, CreateRecordParams{
			IDUser: userID,
			IDGame: gameID,
			Score:  sql.NullInt32{Int32: 2, Valid: true},
			Status: "Completado",
		})
		if err != nil {
			t.Fatalf("Error al crear record: %v\n", err)
		}
	})
	if !success {
		t.Fatalf("No se pudo crear el record, abortando\n")
	}

	t.Run("GetRecord", func(t *testing.T) {
		r, err := queries.GetRecord(ctx, GetRecordParams{
			IDUser: userID,
			IDGame: gameID,
		})
		if err != nil {
			t.Fatalf("Error al obtener record: %v\n", err)
		}
		fmt.Printf("IDs del registro: %d, %d\n", r.IDUser, r.IDGame)
		if r.Status != "Completado" {
			t.Errorf("Status esperado %q, obtenido %q\n", "Completado", r.Status)
		} else {
			fmt.Printf("Status del registro: %s\n", r.Status)
		}
		if !r.Score.Valid || r.Score.Int32 != 2 {
			t.Errorf("Score esperado %q, obtenido %q\n", 2, r.Score.Int32)
		} else {
			fmt.Printf("Score: %d\n", r.Score.Int32)
		}
	})

	t.Run("UpdateRecord", func(t *testing.T) {
		err := queries.UpdateRecord(ctx, UpdateRecordParams{
			IDUser: userID,
			IDGame: gameID,
			Score:  sql.NullInt32{Int32: 4, Valid: true},
			Status: "Platinado",
		})
		if err != nil {
			t.Fatalf("Error al actualizar record: %v\n", err)
		}
		// Leer actualizado
		r, err := queries.GetRecord(ctx, GetRecordParams{IDUser: userID, IDGame: gameID})
		if err != nil {
			t.Fatalf("Error al releer record: %v\n", err)
		}
		fmt.Printf("IDs del registro: %d, %d\n", r.IDUser, r.IDGame)
		if r.Status != "Platinado" {
			t.Errorf("Status esperado %q, obtenido %q\n", "Platinado", r.Status)
		} else {
			fmt.Printf("Status del registro: %s\n", r.Status)
		}
		fmt.Printf("Score: %d\n", r.Score.Int32)
	})

	t.Run("ListRecord", func(t *testing.T) {
		records, err := queries.LisRecord(ctx, userID)
		if err != nil {
			t.Fatalf("Error al listar records: %v", err)
		}
		found := false
		fmt.Println("-----LISTADO-----")
		for _, r := range records {
			fmt.Printf("IDs: %v, %v  Status: %s  Puntuación: %d\n", r.IDUser, r.IDGame, r.Status, r.Score.Int32)
			fmt.Println("--------")
			if r.IDUser == userID && r.IDGame == gameID {
				found = true
			}
		}
		if !found {
			t.Errorf("El registro creado (IDs %d, %d) no aparece en ListRecord\n", userID, gameID)
		}
	})

	// ---------- Cleanup: Record, Game y User ----------
	t.Run("DeleteRecord", func(t *testing.T) {
		err := queries.DeleteRecord(ctx, DeleteRecordParams{
			IDUser: userID,
			IDGame: gameID,
		})
		if err != nil {
			t.Fatalf("Error al eliminar record: %v\n", err)
		}
		_, err = queries.GetRecord(ctx, GetRecordParams{IDUser: userID, IDGame: gameID})
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("Se esperaba sql.ErrNoRows al buscar el registro eliminado, se obtuvo: %v\n", err)
		} else {
			fmt.Printf("El registro con IDs %d, %d no existe\n", userID, gameID)
		}
	})

	t.Run("DeleteGame", func(t *testing.T) {
		err := queries.DeleteGame(ctx, gameID)
		if err != nil {
			t.Fatalf("Error al eliminar juego: %v\n", err)
		}
		_, err = queries.GetGame(ctx, gameID)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("Se esperaba sql.ErrNoRows al buscar el juego eliminado, se obtuvo: %v\n", err)
		} else {
			fmt.Printf("El juego con ID %d no existe\n", gameID)
		}
	})

	t.Run("DeleteUser", func(t *testing.T) {
		err := queries.DeleteUser(ctx, userID)
		if err != nil {
			t.Fatalf("Error al eliminar usuario: %v\n", err)
		}
		// Try to fetch User → fail
		_, err = queries.GetUser(ctx, userID)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("Se esperaba sql.ErrNoRows al buscar el usuario eliminado, se obtuvo: %v\n", err)
		} else {
			fmt.Printf("El usuario con ID %d no existe\n", userID)
		}
	})
}
