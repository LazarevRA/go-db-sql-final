package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	db, err := sql.Open("sqlite", "tracker.db") // настройка подключения к БД

	require.NoError(t, err)

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	number, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEqual(t, 0, number)

	parcel.Number = number // присваиваем номер тестовой посылке.

	// get
	// получаем только что добавленную посылку, проверяем, что нет ошибки
	// проверяем, что значения всех полей в проверяемой посылке совпадают с тестовой
	parcelToChech, err := store.Get(number)
	require.NoError(t, err)
	assert.Equal(t, parcel, parcelToChech)

	// delete
	// удаляем добавленную посылку, проверяем отсутствие ошибки
	// проверяем, что посылку больше нельзя получить из БД
	err = store.Delete(number)
	require.NoError(t, err)

	_, err = store.Get(number)
	require.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройка подключения к БД

	require.NoError(t, err)

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавляем новую посылку в БД, проверяем отсутствие ошибки и наличие идентификатора
	number, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEqual(t, 0, number)

	// set address
	// обновляем адрес, проверяем отсутствие ошибки
	newAddress := "new test address"
	err = store.SetAddress(number, newAddress)
	require.NoError(t, err)

	// check
	// получаем добавленную посылку и полверяем, что адрес обновился и нет ошибки
	parcelToChech, err := store.Get(number)
	require.NoError(t, err)
	assert.Equal(t, newAddress, parcelToChech.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройка подключения к БД

	require.NoError(t, err)

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавляем новую посылку в БД, проверяем отсутствие ошибки и наличие идентификатора
	number, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEqual(t, 0, number)

	// set status
	// обновляем статус, проверяем отсутствие ошибки
	err = store.SetStatus(number, ParcelStatusDelivered)
	require.NoError(t, err)

	// check
	// получаем добавленную посылку и полверяем, что статус обновился и нет ошибки
	parcelToChech, err := store.Get(number)
	require.NoError(t, err)
	assert.Equal(t, ParcelStatusDelivered, parcelToChech.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройка подключения к БД

	require.NoError(t, err)

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
		require.NoError(t, err)
		require.NotEqual(t, 0, id)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	// получаем список посылок по идентификатору клиента, сохранённого в переменной client
	storedParcels, err := store.GetByClient(client)
	// проверяем отсутствие ошибки
	require.NoError(t, err)
	// проверяем, что количество полученных посылок совпадает с количеством добавленных
	assert.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// проверяем, что все посылки из storedParcels есть в parcelMap
		_, ok := parcelMap[parcel.Number]
		assert.True(t, ok)
		// убедитесь, что значения полей полученных посылок заполнены верно
		assert.Equal(t, parcel, parcelMap[parcel.Number])
	}
}
