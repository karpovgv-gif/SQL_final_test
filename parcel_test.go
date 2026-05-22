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
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		return
	}
	defer db.Close()
	// настройте подключение к БД

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	res, err := store.Add(parcel)

	require.NoError(t, err, "Ошибка при добавлении посылки")
	assert.NotEmpty(t, res, "Отсутствует идентификатор посылки")

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	getParcel, err := store.Get(res)

	require.NoError(t, err, "Ошибка при получении посылки")

	assert.Equal(t, getParcel.Client, parcel.Client)
	assert.Equal(t, getParcel.Status, parcel.Status)
	assert.Equal(t, getParcel.Address, parcel.Address)
	assert.Equal(t, getParcel.CreatedAt, parcel.CreatedAt)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(res)

	require.NoError(t, err, "Ошибка при удалении посылки")

	_, err = store.Get(res)
	require.Error(t, err)

}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		return
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	res, err := store.Add(parcel)

	require.NoError(t, err)
	assert.NotEmpty(t, res)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(res, newAddress)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	newParcel, err := store.Get(res)

	require.NoError(t, err)
	assert.Equal(t, newParcel.Address, newAddress)

}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		return
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	res, err := store.Add(parcel)

	require.NoError(t, err)
	assert.NotEmpty(t, res)

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(res, ParcelStatusSent)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	newParcel, err := store.Get(res)

	require.NoError(t, err)
	assert.Equal(t, newParcel.Status, ParcelStatusSent)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		return
	}
	defer db.Close()

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

	store := NewParcelStore(db)

	// add
	//parcel := getTestParcel()
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])

		require.NoError(t, err)
		assert.NotEmpty(t, id) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	// получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	storedParcels, err := store.GetByClient(client)
	assert.Equal(t, len(parcels), len(storedParcels))
	require.NoError(t, err)

	// check
	// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
	// убедитесь, что все посылки из storedParcels есть в parcelMap
	// убедитесь, что значения полей полученных посылок заполнены верно
	for _, parcel := range storedParcels {
		expected, ok := parcelMap[parcel.Number]
		require.True(t, ok)

		assert.Equal(t, expected.Number, parcel.Number)
		assert.Equal(t, expected.Client, parcel.Client)
		assert.Equal(t, expected.Status, parcel.Status)
		assert.Equal(t, expected.Address, parcel.Address)
		assert.Equal(t, expected.CreatedAt, parcel.CreatedAt)
	}
}
