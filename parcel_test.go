package main

import (
	"database/sql"
	"errors"
	"math/rand"
	"testing"
	"time"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	if err != nil {
		t.Fatalf("не удалось добавить строку в таблицу: %v", err)
	}
	parcel.Number = id
	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	got, err := store.Get(id)
	if err != nil {
		t.Fatalf("не удалось прочитать строку: %v", err)
	}
	if got.Client != parcel.Client || got.Status != parcel.Status || got.Address != parcel.Address {
		t.Errorf("поля полученной строки отличаются от поле в parsel: got %+v, want %+v", got, parcel)
	}
	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(id)
	if err != nil {
		t.Fatalf("ошибка в удалении посылки: %v", err)
	}

	_, err = store.Get(id)
	if err == nil {
		t.Fatalf("посылка все еще существует, ошибка удаления")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf(" ожидалась ошибка sql.ErrNoRows, результат %v", err)
	}
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	if err != nil {
		t.Fatalf("не удалось добавить строку в таблицу: %v", err)
	}
	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	if err != nil {
		t.Fatalf("ошибка в обновлении адреса: %v", err)
	}

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	got, err := store.Get(id)
	if err != nil {
		t.Fatalf("не удалось получить строку: %v", err)
	}

	if got.Address != newAddress {
		t.Errorf("адрес не обновился: got %q, want %q", got.Address, newAddress)
	}
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	if err != nil {
		t.Fatalf("не удалось добавить строку в таблицу: %v", err)
	}

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	newStatus := "sent"
	err = store.SetStatus(id, newStatus)
	if err != nil {
		t.Fatalf("не удалось обновить статус: %v", err)
	}

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	got, err := store.Get(id)
	if err != nil {
		t.Fatalf("не удалось получить строку: %v", err)
	}

	if got.Status != newStatus {
		t.Errorf("статус не обновился: got %q, want %q", got.Status, newStatus)
	}
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i]) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		if err != nil {
			t.Fatalf("не удалось добавить строку в таблицу: %v", err)
		}
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	if err != nil {
		t.Fatalf("не удалось получить список посылок по идентификатору: %v", err)
	}

	if len(storedParcels) != len(parcels) {
		t.Fatalf("ожидалось %d посылок, получено %d", len(parcels), len(storedParcels))
	}

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		orig, ok := parcelMap[parcel.Number]
		if !ok {
			t.Errorf("посылка с номером %d не найдена из перечня добавленных", parcel.Number)
			continue
		}
		if parcel.Client != orig.Client || parcel.Status != orig.Status || parcel.Address != orig.Address {
			t.Errorf("несоответствие %d: got %+v, want %+v", parcel.Number, parcel, orig)
		}
	}
}
