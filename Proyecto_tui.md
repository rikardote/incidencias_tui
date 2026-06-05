# Proyecto TUI / API Standalone para Incidencias

## Avance actual — 2026-06-04

### Backend Laravel (`incidencias_modern`)

Rama de trabajo:

```txt
proyecto-tui-api
```

Estado:

- API `/api/v1` creada y protegida con Sanctum.
- Auth lista: login, logout, usuario actual.
- Catálogos listos: empleados, códigos, médicos, periodos, qnas y departamentos.
- API de incidencias lista: captura y eliminación por token usando `IncidenciasService`.
- API biométrica lista: recientes, asistencia por empleado/departamento, dispositivos y sync admin.
- API de reportes lista: recientes, empleado, resumen qna, RH5, estadísticas y ausentismo.
- Documentadas pruebas Docker con puerto `8190`.

Commits relevantes:

```txt
4cdf99e feat: add v1 api auth and catalogs for TUI
56b87e2 feat: add biometric api endpoints
065c623 feat: add incidencias capture api
09f05dd feat: add reports api endpoints
786e15c docs: add docker api testing notes
```

Prueba validada:

```txt
GET http://localhost:8190/api/v1/employees?search=332618
```

Resultado validado:

```txt
332618 - HECTOR RICARDO FUENTES ARMENTA
```

### Cliente TUI standalone (`~/Documents/code/incidencias_tui`)

Estado:

- Proyecto Go creado con Bubble Tea.
- Git inicializado en rama `main`.
- Login funcional contra `POST /api/v1/login`.
- Menú principal funcional.
- Búsqueda de empleados funcional contra `GET /api/v1/employees?search=`.
- Navegación de campos de login corregida:
  - `Tab`, `↓`, `Ctrl+N`: siguiente campo.
  - `Shift+Tab`, `↑`, `Ctrl+P`: campo anterior.
  - `Enter` en usuario pasa a contraseña.
  - `Enter` en contraseña hace login.
- Placeholders pendientes para captura, reportes y biométrico.
- Puede ejecutarse sin Go local usando Docker:

```bash
cd ~/Documents/code/incidencias_tui
make docker-run
```

Commits locales del TUI:

```txt
df63e1c feat: scaffold incidencias TUI client
724a223 docs: add docker run instructions
63c4220 chore: add docker make targets
6e8bca8 fix: improve login field navigation
```

### Pendiente para la siguiente sesión

1. Crear repositorio remoto GitHub para `incidencias_tui`.
2. Subir commits del TUI.
3. Implementar pantalla de selección/búsqueda de empleado más completa.
4. Implementar catálogos en TUI: códigos, periodos, médicos.
5. Implementar captura de incidencias en TUI.
6. Implementar reportes básicos en TUI.
7. Implementar sección biométrica en TUI.
8. Agregar persistencia segura de token en `~/.config/incidencias-tui/config.json`.
9. Preparar build/release binario.

---

## Objetivo

Crear una base API en Laravel que permita consumir el sistema de incidencias desde clientes externos, inicialmente una aplicación **TUI standalone**, y posteriormente una posible **app Android** u otros clientes.

La aplicación web actual con Livewire debe seguir funcionando igual. La nueva API debe reutilizar la lógica existente del proyecto, especialmente los servicios de dominio, para evitar duplicar reglas.

---

## Arquitectura propuesta

```txt
Laravel actual
├── Web Livewire existente
├── API JSON /api/v1
│   ├── TUI standalone
│   ├── App Android
│   └── otros clientes futuros
└── Base de datos / reglas existentes
```

El cliente TUI no debe conectarse directamente a la base de datos. Debe consumir Laravel mediante HTTP/JSON.

---

## Principios importantes

1. Laravel sigue siendo el backend central.
2. El TUI solo será una interfaz cliente.
3. La captura debe usar `App\Services\Incidencias\IncidenciasService`.
4. No duplicar reglas de negocio en el TUI ni en Android.
5. Usar autenticación por token.
6. Diseñar endpoints generales, no específicos de TUI.
7. Nombrar rutas como `/api/v1/...` en lugar de `/api/tui/...`.

---

## Autenticación recomendada

Usar **Laravel Sanctum** para tokens personales de API.

Flujo:

```txt
Cliente envía usuario/contraseña
↓
POST /api/v1/login
↓
Laravel valida credenciales
↓
Laravel devuelve token
↓
Cliente usa Authorization: Bearer TOKEN
```

Headers para peticiones autenticadas:

```http
Authorization: Bearer {token}
Accept: application/json
Content-Type: application/json
```

---

## Endpoints mínimos necesarios

### Auth

```txt
POST /api/v1/login
POST /api/v1/logout
GET  /api/v1/me
```

### Catálogos y búsqueda

```txt
GET /api/v1/employees?search=
GET /api/v1/incidence-codes?search=
GET /api/v1/doctors?search=
GET /api/v1/periodos
GET /api/v1/qnas
GET /api/v1/departments
```

### Captura

```txt
POST /api/v1/incidencias
DELETE /api/v1/incidencias/{token}
```

### Reportes

```txt
GET /api/v1/reports/recent
GET /api/v1/reports/employee/{employee}
GET /api/v1/reports/qna-summary
GET /api/v1/reports/rh5
GET /api/v1/reports/statistics
GET /api/v1/reports/absenteeism
```

### Biométrico

```txt
GET  /api/v1/biometrico/recent
GET  /api/v1/biometrico/employee/{employee}/attendance
GET  /api/v1/biometrico/department-attendance
GET  /api/v1/biometrico/devices
POST /api/v1/biometrico/sync
```

---

## Detalle de endpoints

## 1. Login

### `POST /api/v1/login`

Request:

```json
{
  "username": "usuario",
  "password": "contraseña",
  "device_name": "incidencias-tui"
}
```

Response:

```json
{
  "token": "plain-text-token",
  "user": {
    "id": 1,
    "name": "Usuario",
    "username": "usuario",
    "type": "admin",
    "can_capture": true
  }
}
```

Validaciones:

- Usuario y contraseña requeridos.
- Usuario activo.
- Respetar usuarios del modelo `App\Models\User`.

---

## 2. Usuario actual

### `GET /api/v1/me`

Response:

```json
{
  "id": 1,
  "name": "Usuario",
  "username": "usuario",
  "type": "admin",
  "can_capture": true,
  "is_admin": true,
  "departments": []
}
```

---

## 3. Buscar empleados

### `GET /api/v1/employees?search=perez`

Debe respetar permisos por departamento:

- Admin ve todos.
- Member/consulta solo ve empleados de sus departamentos.

Response:

```json
{
  "data": [
    {
      "id": 123,
      "num_empleado": "001234",
      "full_name": "NOMBRE APELLIDO APELLIDO",
      "department": {
        "id": 5,
        "code": "00005",
        "description": "DEPARTAMENTO"
      },
      "puesto": "PUESTO",
      "horario": "HORARIO",
      "jornada": "JORNADA"
    }
  ]
}
```

---

## 4. Códigos de incidencia

### `GET /api/v1/incidence-codes?search=vacaciones`

Response sugerido:

```json
{
  "data": [
    {
      "id": 10,
      "code": "60",
      "description": "VACACIONES",
      "requires_range": true,
      "requires_medico": false,
      "requires_periodo": true,
      "requires_txt": false,
      "requires_comision": false,
      "requires_otorgado": false,
      "is_incapacidad": false,
      "is_licencia": false,
      "is_vacacional": true
    }
  ]
}
```

Esto ayuda al TUI y Android a saber qué campos mostrar.

---

## 5. Médicos

### `GET /api/v1/doctors?search=lopez`

Usar la misma lógica de puestos médicos que existe actualmente en Livewire.

Response:

```json
{
  "data": [
    {
      "id": 55,
      "num_empleado": "005555",
      "full_name": "DR NOMBRE APELLIDOS"
    }
  ]
}
```

---

## 6. Periodos vacacionales

### `GET /api/v1/periodos`

Response:

```json
{
  "data": [
    {
      "id": 1,
      "periodo": 1,
      "year": 2026,
      "label": "Periodo 1/2026"
    }
  ]
}
```

---

## 7. Quincenas

### `GET /api/v1/qnas`

Response:

```json
{
  "data": [
    {
      "id": 100,
      "qna": "11",
      "year": 2026,
      "description": "QNA 11/2026",
      "active": true,
      "cierre": "2026-06-15"
    }
  ]
}
```

---

## 8. Departamentos

### `GET /api/v1/departments`

Debe respetar permisos:

- Admin ve todos.
- Otros usuarios solo sus departamentos asignados.

Response:

```json
{
  "data": [
    {
      "id": 5,
      "code": "00005",
      "description": "DEPARTAMENTO"
    }
  ]
}
```

---

## 9. Capturar incidencia

### `POST /api/v1/incidencias`

Request general:

```json
{
  "employee_id": 123,
  "codigo": 10,
  "fecha_inicio": "2026-06-01",
  "fecha_final": "2026-06-01",

  "medico_id": null,
  "fecha_expedida": null,
  "diagnostico": null,
  "num_licencia": null,

  "periodo_id": null,

  "autoriza_txt": null,
  "cobertura_txt": null,

  "motivo_comision": null,
  "otorgado": null,

  "saltar_validacion_inca": false,
  "saltar_validacion_lic": false
}
```

Internamente debe mapearse a lo que espera `IncidenciasService`:

```php
[
    'empleado_id' => $request->employee_id,
    'codigo' => $request->codigo,
    'datepicker_inicial' => $request->fecha_inicio,
    'datepicker_final' => $request->fecha_final,
    'datepicker_expedida' => $request->fecha_expedida,
    'medico_id' => $request->medico_id,
    'diagnostico' => $request->diagnostico,
    'num_licencia' => $request->num_licencia,
    'periodo_id' => $request->periodo_id,
    'autoriza_txt' => $request->autoriza_txt,
    'cobertura_txt' => $request->cobertura_txt,
    'motivo_comision' => $request->motivo_comision,
    'otorgado' => $request->otorgado,
    'saltar_validacion_inca' => $request->boolean('saltar_validacion_inca') ? '1' : '0',
    'saltar_validacion_lic' => $request->boolean('saltar_validacion_lic') ? '1' : '0',
    'token' => $token,
]
```

Response exitoso:

```json
{
  "message": "Incidencia capturada correctamente",
  "token": "batch-token",
  "employee_id": 123
}
```

Response con error de negocio:

```json
{
  "message": "Incidencia Duplicada"
}
```

Código HTTP sugerido:

```txt
422 Unprocessable Entity
```

Reglas importantes:

- Verificar `canCapture()`.
- Verificar mantenimiento.
- Verificar acceso al empleado por departamento.
- Admin puede saltar validaciones si aplica.
- Member no debería poder saltar validaciones especiales.
- Llamar a `event(new NewIncidenciaBatchCreated())` después de capturar.

---

## 10. Eliminar incidencia por token

### `DELETE /api/v1/incidencias/{token}`

Debe usar:

```php
App\Services\Incidencias\IncidenciasService::eliminarPorToken($token)
```

Response:

```json
{
  "message": "Incidencia eliminada correctamente",
  "deleted": 2
}
```

---

## 11. Reporte: últimas incidencias

### `GET /api/v1/reports/recent?limit=20`

Response:

```json
{
  "data": [
    {
      "fecha_capturado": "2026-06-01 10:30",
      "employee": {
        "id": 123,
        "num_empleado": "001234",
        "full_name": "NOMBRE APELLIDOS"
      },
      "codigo": {
        "id": 10,
        "code": "60",
        "description": "VACACIONES"
      },
      "fecha_inicio": "2026-06-01",
      "fecha_final": "2026-06-05",
      "total_dias": 5,
      "qna": "Q11/2026",
      "capturado_por": "U."
    }
  ]
}
```

---

## 12. Reporte: historial por empleado

### `GET /api/v1/reports/employee/{employee}?start=2026-01-01&end=2026-12-31`

Response:

```json
{
  "employee": {
    "id": 123,
    "num_empleado": "001234",
    "full_name": "NOMBRE APELLIDOS"
  },
  "data": [
    {
      "codigo": "60",
      "description": "VACACIONES",
      "fecha_inicio": "2026-06-01",
      "fecha_final": "2026-06-05",
      "total_dias": 5,
      "qna": "Q11/2026",
      "periodo": "1/2026"
    }
  ]
}
```

---

## 13. Reporte: resumen por quincena

### `GET /api/v1/reports/qna-summary?qna_id=100&department_id=5`

Response:

```json
{
  "data": [
    {
      "code": "60",
      "description": "VACACIONES",
      "registros": 12,
      "dias": 45
    }
  ]
}
```

### Reportes adicionales API

```txt
GET /api/v1/reports/rh5?qna_id=100&department_id=5
GET /api/v1/reports/statistics?code_id=10&start=2026-01-01&end=2026-12-31&department_id=5&gender=H
GET /api/v1/reports/absenteeism?start=2026-01-01&end=2026-12-31&department_id=5
```

Estos endpoints permiten cubrir reportes base para TUI/Android sin generar PDF/Excel.

---

## 14. Biométrico

Estos endpoints sirven para la sección biométrica del TUI, Android u otros clientes. El API WearOS existente puede mantenerse, pero estos endpoints quedan protegidos con token Sanctum y permisos del sistema.

### `GET /api/v1/biometrico/recent?limit=50`

Devuelve las últimas checadas visibles para el usuario autenticado.

Response:

```json
{
  "data": [
    {
      "id": 1,
      "num_empleado": "001234",
      "employee": {
        "id": 123,
        "num_empleado": "001234",
        "full_name": "NOMBRE APELLIDOS"
      },
      "fecha": "2026-06-04 08:00:00",
      "hora": "08:00:00",
      "timestamp": 1780560000000,
      "identificador": "001234_202606040800_ADMS",
      "location": "ADMS_Entrada"
    }
  ]
}
```

### `GET /api/v1/biometrico/employee/{employee}/attendance`

Parámetros posibles:

```txt
?start=2026-06-01&end=2026-06-15
```

O por quincena:

```txt
?year=2026&qna=11&qna_end=12
```

Response:

```json
{
  "employee": {},
  "start": "2026-06-01",
  "end": "2026-06-15",
  "data": [
    {
      "date": "2026-06-01",
      "primera_checada": "2026-06-01 08:00:00",
      "ultima_checada": "2026-06-01 15:00:00",
      "hora_entrada": "08:00:00",
      "hora_salida": "15:00:00",
      "num_checadas": 2,
      "retardo": false,
      "incidencias": [],
      "incidencias_tokens": []
    }
  ]
}
```

### `GET /api/v1/biometrico/department-attendance`

Parámetros:

```txt
?department_id=5&year=2026&qna=11
```

Devuelve asistencia agrupada por empleado para un departamento permitido.

### `GET /api/v1/biometrico/devices`

Solo administradores. Lista equipos biométricos configurados.

### `POST /api/v1/biometrico/sync`

Solo administradores. Sincroniza equipos biométricos.

Request opcional:

```json
{
  "device_ids": [1, 2]
}
```

Si no se envía `device_ids`, intenta sincronizar todos los equipos.

---

## Estructura Laravel sugerida

Crear controladores API:

```txt
app/Http/Controllers/Api/V1/AuthController.php
app/Http/Controllers/Api/V1/EmployeeController.php
app/Http/Controllers/Api/V1/CatalogController.php
app/Http/Controllers/Api/V1/BiometricoController.php
app/Http/Controllers/Api/V1/IncidenciaController.php
app/Http/Controllers/Api/V1/ReportController.php
```

Crear requests:

```txt
app/Http/Requests/Api/V1/LoginRequest.php
app/Http/Requests/Api/V1/StoreIncidenciaRequest.php
```

Crear resources opcionalmente:

```txt
app/Http/Resources/Api/V1/EmployeeResource.php
app/Http/Resources/Api/V1/IncidenceCodeResource.php
app/Http/Resources/Api/V1/IncidenciaResource.php
```

Rutas:

```php
Route::prefix('v1')->group(function () {
    Route::post('/login', [AuthController::class, 'login']);

    Route::middleware('auth:sanctum')->group(function () {
        Route::post('/logout', [AuthController::class, 'logout']);
        Route::get('/me', [AuthController::class, 'me']);

        Route::get('/employees', [EmployeeController::class, 'index']);
        Route::get('/incidence-codes', [CatalogController::class, 'incidenceCodes']);
        Route::get('/doctors', [CatalogController::class, 'doctors']);
        Route::get('/periodos', [CatalogController::class, 'periodos']);
        Route::get('/qnas', [CatalogController::class, 'qnas']);
        Route::get('/departments', [CatalogController::class, 'departments']);

        Route::post('/incidencias', [IncidenciaController::class, 'store']);
        Route::delete('/incidencias/{token}', [IncidenciaController::class, 'destroy']);

        Route::get('/reports/recent', [ReportController::class, 'recent']);
        Route::get('/reports/employee/{employee}', [ReportController::class, 'employee']);
        Route::get('/reports/qna-summary', [ReportController::class, 'qnaSummary']);
    });
});
```

---

## Seguridad y permisos

La API debe validar:

1. Token válido.
2. Usuario activo.
3. Tipo de usuario.
4. Permisos de captura.
5. Acceso a empleados por departamento.
6. Modo mantenimiento.
7. Quincenas cerradas.
8. Captura en quincenas cerradas solo con excepción activa.
9. No exponer datos sensibles innecesarios.
10. Rate limiting en login.

---

## Cliente TUI standalone

Recomendación tecnológica:

```txt
Go + Bubble Tea + Bubbles + Lip Gloss
```

Ventajas:

- Binario standalone.
- No necesita PHP local.
- No depende de Docker.
- Funciona en Linux/macOS/Windows.
- Buena experiencia terminal.

Comando objetivo:

```bash
./incidencias-tui
```

Configuración:

```env
INCIDENCIAS_API_URL=https://sistema.example.com
```

Pantallas mínimas:

1. Login.
2. Menú principal.
3. Búsqueda de empleado.
4. Selección de código.
5. Formulario dinámico según código.
6. Confirmación de captura.
7. Reportes.
8. Logout.

---

## App Android futura

Con la misma API también se podría crear una app Android.

Opciones:

1. Kotlin nativo.
2. Flutter.
3. React Native.

La app Android consumiría los mismos endpoints `/api/v1`.

No debería implementar reglas de negocio localmente; solo mostrar formularios, enviar datos y mostrar errores devueltos por Laravel.

---

## Pruebas en Docker

La app corre detrás de Nginx en el puerto local:

```txt
http://localhost:8190
```

Para preparar el contenedor después de traer la rama:

```bash
docker compose exec app composer install
docker compose exec app php artisan migrate --force
docker compose exec app php artisan optimize:clear
```

Si solo falta la tabla de tokens Sanctum:

```bash
docker compose exec app php artisan migrate --path=database/migrations/2026_06_04_000001_create_personal_access_tokens_table.php --force
```

Para listar las rutas API dentro del contenedor usar:

```bash
docker compose exec app php artisan route:list --path=v1
```

Nota: `--path=api/v1` puede no mostrar resultados dependiendo de cómo Laravel compare el filtro de rutas. Con `--path=v1` sí deben aparecer rutas como:

```txt
api/v1/login
api/v1/employees
api/v1/incidencias
api/v1/biometrico/recent
api/v1/reports/recent
```

### Ejemplo: consultar empleado 332618

1. Obtener token con login:

```bash
curl -X POST http://localhost:8190/api/v1/login \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "TU_USUARIO",
    "password": "TU_PASSWORD",
    "device_name": "test-api"
  }'
```

2. Consultar empleado:

```bash
curl "http://localhost:8190/api/v1/employees?search=332618" \
  -H "Accept: application/json" \
  -H "Authorization: Bearer TU_TOKEN"
```

Respuesta observada en desarrollo:

```json
{
  "data": [
    {
      "id": 1430,
      "num_empleado": "332618",
      "full_name": "HECTOR RICARDO FUENTES ARMENTA",
      "department": {
        "id": 19,
        "code": "00105",
        "description": "SUBDELEGACIÓN DE ADMINISTRACIÓN"
      },
      "puesto": "PROFESIONAL ADMINISTRATIVO \"A\"",
      "horario": "08:00 A 17:00",
      "jornada": "MATUTINO"
    }
  ]
}
```

---

## Fases sugeridas

## Fase 1: API base

- Instalar/configurar Sanctum si no está instalado.
- Crear rutas `/api/v1`.
- Crear login/logout/me.
- Crear endpoints de catálogos.
- Crear endpoint de búsqueda de empleados.

## Fase 2: Captura por API

- Crear `POST /api/v1/incidencias`.
- Reutilizar `IncidenciasService`.
- Validar permisos.
- Validar mantenimiento.
- Probar incapacidades, vacaciones, TXT, comisión y omisiones.

## Fase 3: Reportes API

- Últimas incidencias.
- Historial por empleado.
- Resumen por quincena.

## Fase 4: Cliente TUI

- Crear proyecto Go.
- Implementar cliente HTTP.
- Login/token.
- Pantallas principales.
- Captura.
- Reportes.
- Empaquetado binario.

## Fase 5: Android opcional

- Usar los mismos endpoints.
- Diseñar UI móvil.
- Login/token seguro.
- Captura y reportes.

---

## Pendientes de decisión

1. Confirmar si se usará Sanctum o Passport.
2. Definir URL final de API.
3. Definir si el TUI guardará token local o pedirá login siempre.
4. Definir reportes mínimos para primera versión.
5. Definir si habrá modo offline o solo online.
6. Definir distribución del binario TUI.

---

## Recomendación final

Primero construir `/api/v1` de forma genérica. Después cualquier cliente, TUI o Android, podrá consumirla sin tocar la lógica central del sistema.
