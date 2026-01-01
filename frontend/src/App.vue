<template>
  <div class="visualizer">
    <h1>Go Map Buckets Visualizer</h1>
    <button @click="loadData" :disabled="loading">
      {{ loading ? 'Загрузка...' : 'Обновить данные' }}
    </button>

    <div v-if="error" class="error">{{ error }}</div>

    <div v-if="columns.length === 0 && !loading" class="empty">
      Нет данных. Запустите Go-сервер и добавьте элементы в map.
    </div>

    <div class="grid" v-if="columns.length > 0">
      <div
        v-for="(column, colIndex) in columns"
        :key="colIndex"
        class="column"
      >
        <div
          v-for="(bucket, bucketIndex) in column"
          :key="bucket.id"
          class="bucket"
          :class="{ 'is-overflow': bucket.type === 'overflow' }"
        >
          <!-- Tophash -->
          <div class="tophash-row">
            <div
              v-for="(th, i) in bucket.tophash"
              :key="i"
              class="tophash-cell"
            >
              {{ th }}
            </div>
          </div>

          <!-- Keys (зелёные) -->
          <div class="keys-row">
            <div
              v-for="(key, i) in bucket.Keys"
              :key="i"
              class="cell key-cell"
            >
              {{ key === null ? '<empty>' : formatKey(key) }}
            </div>
          </div>

          <!-- Values (жёлтые) -->
          <div class="values-row">
            <div
              v-for="(val, i) in bucket.Values"
              :key="i"
              class="cell value-cell"
            >
              {{ val === null ? '<empty>' : formatValue(val) }}
            </div>
          </div>

          <!-- Overflow адрес -->
          <div class="overflow">
            {{ bucket.overflow }}
          </div>

          <!-- Стрелка вниз к следующему в цепочке -->
          <div
            v-if="bucketIndex < column.length - 1"
            class="arrow-down"
          >
            ↓
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const columns = ref([])
const loading = ref(false)
const error = ref('')

async function loadData() {
  loading.value = true
  error.value = ''
  columns.value = []

  try {
	const res = await fetch('/vizual')
    if (!res.ok) throw new Error('Сервер не отвечает')
    const data = await res.json()

    const rawBuckets = data.buckets || []

    // Формируем столбцы: main + его overflow подряд
    let currentColumn = []

    rawBuckets.forEach(bucket => {
      if (bucket.type === 'main') {
        if (currentColumn.length > 0) {
          columns.value.push(currentColumn)
        }
        currentColumn = [bucket]
      } else {
        // overflow
        currentColumn.push(bucket)
      }
    })

    if (currentColumn.length > 0) {
      columns.value.push(currentColumn)
    }
  } catch (err) {
    error.value = 'Ошибка загрузки: ' + err.message
    console.error(err)
  } finally {
    loading.value = false
  }
}

function formatKey(key) {
  return key // для int — просто число
}

function formatValue(val) {
  if (typeof val === 'object' && val !== null) {
    // Для структур — показываем кратко
    return JSON.stringify(val, null, 0).slice(0, 40) + (JSON.stringify(val).length > 40 ? '...' : '')
  }
  return val
}

// Автозагрузка при старте
loadData()
</script>

<style scoped>
.visualizer {
  padding: 20px;
  font-family: 'Courier New', Courier, monospace;
  background: #f9f9f9;
  min-height: 100vh;
}

h1 {
  text-align: center;
  color: #2c3e50;
  margin-bottom: 20px;
}

button {
  display: block;
  margin: 20px auto;
  padding: 12px 24px;
  font-size: 16px;
  background: #3498db;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

button:disabled {
  background: #95a5a6;
  cursor: not-allowed;
}

.error {
  color: red;
  text-align: center;
  font-weight: bold;
}

.empty {
  text-align: center;
  color: #7f8c8d;
  font-size: 20px;
  margin-top: 100px;
}

.grid {
  display: flex;
  justify-content: center;
  gap: 80px;
  flex-wrap: wrap;
  margin-top: 40px;
}

.column {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 30px;
}

.bucket {
  width: 720px;
  border: 3px solid #34495e;
  border-radius: 12px;
  background: #fff;
  padding: 15px;
  box-shadow: 0 6px 12px rgba(0,0,0,0.15);
}

.bucket.is-overflow {
  border-color: #e67e22;
  background: #fef9e6;
}

.tophash-row {
  display: flex;
  gap: 4px;
  margin-bottom: 12px;
  justify-content: center;
}

.tophash-cell {
  width: 85px;
  height: 28px;
  background: #bdc3c7;
  border: 1px solid #95a5a6;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: bold;
  color: #2c3e50;
}

.keys-row, .values-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  justify-content: center;
  margin-bottom: 12px;
}

.cell {
  width: 85px;
  height: 50px;
  border: 1px solid #7f8c8d;
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  font-size: 12px;
  padding: 4px;
  box-sizing: border-box;
  overflow: hidden;
  border-radius: 4px;
}

.key-cell {
  background: #a8e6cf; /* мягкий зелёный */
}

.value-cell {
  background: #ffecb3; /* мягкий жёлтый */
}

.overflow {
  text-align: center;
  margin-top: 12px;
  font-family: monospace;
  font-size: 14px;
  color: #7f8c8d;
  background: #ecf0f1;
  padding: 8px;
  border-radius: 6px;
}

.arrow-down {
  font-size: 40px;
  color: #3498db;
  margin: 10px 0;
}
</style>
