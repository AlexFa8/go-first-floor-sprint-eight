package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	res, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :address, :created_at)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt))
	if err != nil {
		return 0, fmt.Errorf("не удалось добавить строку в таблицу")
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("не удалось извлечь идентификатор последней добавленной записи")
	}

	// верните идентификатор последней добавленной записи
	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка
	// заполните объект Parcel данными из таблицы
	p := Parcel{}
	err := s.db.QueryRow("Select number, client, status, address, created_at from parcel where number = ?", number).Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, err // важно вернуть оригинальный sql.ErrNoRows
		}
		return p, fmt.Errorf("не удалось прочитать строку по заданному номеру: %w", err)
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк
	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = ?", client)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать строку по заданному клиенту")
	}
	defer rows.Close()

	var res []Parcel

	for rows.Next() {
		var p Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("не удалось прочитать строку по заданному клиенту")
		}
		res = append(res, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	_, err := s.db.Exec("UPDATE parcel SET status = ? WHERE number = ?", status, number)

	if err != nil {
		return fmt.Errorf("не удалось обновить статус посылки")
	}
	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	var status string
	err := s.db.QueryRow("SELECT status FROM parcel WHERE number = ?", number).Scan(&status)
	if err != nil {
		return fmt.Errorf("не удалось обновить адрес")
	}

	if status != "registered" {
		return fmt.Errorf("адрес можно изменить только для статуса 'registered'")
	}

	_, err = s.db.Exec("UPDATE parcel SET address = ? WHERE number = ?", address, number)
	if err != nil {
		return fmt.Errorf("не удалось обновить адрес")
	}
	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	var status string
	err := s.db.QueryRow("SELECT status FROM parcel WHERE number = ?", number).Scan(&status)
	if err != nil {
		return fmt.Errorf("не удалось получить статус")
	}

	if status != "registered" {
		return nil
	}

	_, err = s.db.Exec("DELETE FROM parcel WHERE number = ?", number)
	if err != nil {
		return fmt.Errorf("не удалось удалить строку")
	}
	return nil
}
