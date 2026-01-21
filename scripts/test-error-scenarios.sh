#!/bin/bash

# ============================================================================
# 🧪 TEST ERROR SCENARIOS - Log-Trace Correlation Practice
# ============================================================================
# Este script contiene diferentes casos de error para practicar debugging
# usando la correlación entre Logs y Traces en Jaeger
#
# 🎯 OBJETIVO DIDÁCTICO:
# Aprender a identificar errores usando:
# 1. Logs estructurados con trace_id
# 2. Traces en Jaeger UI
# 3. Correlación entre ambos para root cause analysis
# ============================================================================

BASE_URL="http://localhost:8080/api/v1/users"

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}🧪 ERROR SCENARIOS TEST SUITE${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# ============================================================================
# 📋 FUNCIÓN HELPER: Ejecutar test y explicar
# ============================================================================
run_test() {
    local test_number=$1
    local test_name=$2
    local description=$3
    local curl_command=$4
    local expected_error=$5
    local debug_steps=$6
    
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}TEST #${test_number}: ${test_name}${NC}"
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${BLUE}📝 DESCRIPCIÓN:${NC}"
    echo "   $description"
    echo ""
    echo -e "${BLUE}🔧 COMANDO:${NC}"
    echo "   $curl_command"
    echo ""
    echo -e "${BLUE}❌ ERROR ESPERADO:${NC}"
    echo "   $expected_error"
    echo ""
    echo -e "${BLUE}🔍 PASOS DE DEBUG:${NC}"
    echo "$debug_steps"
    echo ""
    echo -e "${YELLOW}Presiona ENTER para ejecutar el test...${NC}"
    read -r
    
    # Ejecutar el comando
    echo -e "${GREEN}📤 EJECUTANDO REQUEST...${NC}"
    eval "$curl_command"
    echo ""
    echo ""
    
    echo -e "${YELLOW}🔍 AHORA SIGUE ESTOS PASOS:${NC}"
    echo "   1. Mira los logs en la terminal del servidor"
    echo "   2. Copia el trace_id que aparece en los logs"
    echo "   3. Ve a Jaeger: http://localhost:16686"
    echo "   4. Busca ese trace_id"
    echo "   5. Analiza el error en el trace"
    echo ""
    echo -e "${YELLOW}Presiona ENTER cuando hayas analizado en Jaeger...${NC}"
    read -r
    echo ""
}

# ============================================================================
# TEST 1: Email ya existe (409 Conflict)
# ============================================================================
# 🎓 CONCEPTO: Error de negocio - violación de constraint único
# Este es el error MÁS COMÚN en APIs REST
# ============================================================================
run_test "1" \
    "Email Already Exists (409 Conflict)" \
    "Intentar crear un usuario con un email que ya existe en la base de datos" \
    "curl -X POST $BASE_URL \
  -H 'Content-Type: application/json' \
  -d '{\"name\":\"Duplicate User\",\"email\":\"tracetest2@example.com\",\"password\":\"password12345\"}'" \
    "409 Conflict - Email already exists" \
    "   a) En logs: Busca 'Failed to create user' con trace_id
   b) En Jaeger: 
      - Ve al span 'UserRepository.GetByEmail'
      - Verá que user.found=true (el email existe)
      - Ve al span 'UserRepository.Create'
      - Verá el atributo error.type='unique_violation'
   c) Conclusión: El email ya existía antes de intentar crear"

# ============================================================================
# TEST 2: Password muy corto (400 Bad Request)
# ============================================================================
# 🎓 CONCEPTO: Error de validación - el middleware lo detecta ANTES del handler
# Este error NO llegará a generar logs en CreateUser porque falla antes
# ============================================================================
run_test "2" \
    "Password Too Short (400 Bad Request)" \
    "Intentar crear usuario con password menor a 8 caracteres" \
    "curl -X POST $BASE_URL \
  -H 'Content-Type: application/json' \
  -d '{\"name\":\"Short Pass\",\"email\":\"shortpass@example.com\",\"password\":\"123\"}'" \
    "400 Bad Request - Password validation failed on 'min' tag" \
    "   a) En logs: NO verás 'Creating user' porque el middleware rechaza antes
   b) Verás el error de validación del middleware
   c) En Jaeger:
      - Verás el span HTTP con status=400
      - NO verás spans de Service/Repository (nunca se ejecutaron)
   d) Conclusión: Validación falla ANTES de la lógica de negocio"

# ============================================================================
# TEST 3: Email inválido (400 Bad Request)
# ============================================================================
# 🎓 CONCEPTO: Validación de formato - middleware valida estructura del email
# ============================================================================
run_test "3" \
    "Invalid Email Format (400 Bad Request)" \
    "Intentar crear usuario con email sin formato válido" \
    "curl -X POST $BASE_URL \
  -H 'Content-Type: application/json' \
  -d '{\"name\":\"Bad Email\",\"email\":\"not-an-email\",\"password\":\"password12345\"}'" \
    "400 Bad Request - Email validation failed on 'email' tag" \
    "   a) Similar al TEST 2, el middleware valida y rechaza
   b) En logs: NO verás logs del handler
   c) En Jaeger: Solo span HTTP con status=400
   d) Conclusión: Validación de estructura antes de negocio"

# ============================================================================
# TEST 4: Nombre vacío (400 Bad Request)
# ============================================================================
# 🎓 CONCEPTO: Campo requerido faltante
# ============================================================================
run_test "4" \
    "Empty Name (400 Bad Request)" \
    "Intentar crear usuario sin nombre" \
    "curl -X POST $BASE_URL \
  -H 'Content-Type: application/json' \
  -d '{\"name\":\"\",\"email\":\"noname@example.com\",\"password\":\"password12345\"}'" \
    "400 Bad Request - Name validation failed on 'required' tag" \
    "   a) Middleware valida que 'name' no esté vacío
   b) En logs: NO verás logs del handler
   c) En Jaeger: Solo span HTTP con status=400
   d) Conclusión: Validación de campos requeridos"

# ============================================================================
# TEST 5: JSON malformado (400 Bad Request)
# ============================================================================
# 🎓 CONCEPTO: Error de parsing - Fiber no puede parsear el JSON
# Este error ocurre AÚN ANTES del middleware de validación
# ============================================================================
run_test "5" \
    "Malformed JSON (400 Bad Request)" \
    "Enviar JSON con sintaxis incorrecta" \
    "curl -X POST $BASE_URL \
  -H 'Content-Type: application/json' \
  -d '{\"name\":\"Test\",\"email\":\"test@test.com\",\"password\":\"pass123\"' \
  # ↑ Falta el } de cierre" \
    "400 Bad Request - Cannot parse JSON" \
    "   a) Error de nivel HTTP, antes de cualquier lógica
   b) En logs: Posible error de parsing
   c) En Jaeger: Span HTTP con error
   d) Conclusión: Error de protocolo/formato"

# ============================================================================
# TEST 6: Usuario no existe (404 Not Found) - GET
# ============================================================================
# 🎓 CONCEPTO: Recurso no encontrado
# Usamos GET para provocar un 404
# ============================================================================
run_test "6" \
    "User Not Found (404 Not Found)" \
    "Intentar obtener un usuario que no existe" \
    "curl -X GET $BASE_URL/00000000-0000-0000-0000-000000000000" \
    "404 Not Found - User not found" \
    "   a) En logs: Verás logs con trace_id del GET handler
   b) En Jaeger:
      - Span HTTP GET /api/v1/users/:id
      - Span UserService.GetUserByID
      - Span UserRepository.GetByID
      - Atributo user.found=false
   c) Conclusión: Query ejecutada correctamente pero sin resultados"

# ============================================================================
# TEST 7: UUID inválido (400 Bad Request) - GET
# ============================================================================
# 🎓 CONCEPTO: Parámetro de ruta con formato inválido
# ============================================================================
run_test "7" \
    "Invalid UUID Format (400 Bad Request)" \
    "Intentar obtener usuario con ID que no es UUID válido" \
    "curl -X GET $BASE_URL/not-a-valid-uuid" \
    "400/500 - Depends on validation" \
    "   a) Depende de si validas el UUID antes de la query
   b) En Jaeger: 
      - Puede fallar en validación (400)
      - O en query de DB (500)
   c) Conclusión: Importancia de validar inputs temprano"

# ============================================================================
# 🎯 EJERCICIO FINAL: Análisis Comparativo
# ============================================================================
echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}🎓 EJERCICIO FINAL: Análisis Comparativo${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""
echo "Ahora que ejecutaste todos los tests, responde:"
echo ""
echo "1️⃣ ¿Cuáles errores generaron logs en CreateUser?"
echo "   → Respuesta: Solo TEST 1 (email duplicado)"
echo "   → Razón: Los demás fallaron en validación antes del handler"
echo ""
echo "2️⃣ ¿Cuáles errores tienen spans de Service/Repository en Jaeger?"
echo "   → Respuesta: TEST 1 y TEST 6"
echo "   → Razón: Solo estos llegaron a ejecutar lógica de negocio"
echo ""
echo "3️⃣ ¿Qué diferencia hay entre TEST 1 y TEST 6?"
echo "   → TEST 1: Error en Create (violación de constraint)"
echo "   → TEST 6: No error en query, simplemente no hay datos"
echo ""
echo "4️⃣ ¿Por qué los TEST 2-5 NO tienen trace_id en logs de handler?"
echo "   → Porque el middleware de validación rechaza la request"
echo "   → El handler CreateUser nunca se ejecuta"
echo "   → Por lo tanto, logger.WithTraceContext() nunca se llama"
echo ""
echo "5️⃣ ¿Cómo mejorarías el debugging para TEST 2-5?"
echo "   → Agregar logger.WithTraceContext() en el middleware de validación"
echo "   → Así los errores de validación también tendrían trace_id"
echo ""
echo -e "${GREEN}✅ ¡Tests completados!${NC}"
echo -e "${YELLOW}Ahora entiendes cómo correlacionar logs y traces para debugging${NC}"
