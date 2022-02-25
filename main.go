package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"time"
)

type Datos struct {
	id_expediente   string
	IdTipoArchivo   int
	hoy             string
	tiempo          string
	usuarioCarga    string
	path            string
	text            string
	descripcion     string
	id_perfil_carga int
	id_busqueda     int
	estatus         int
}

func conexion() (db *sql.DB, e error) {
	fmt.Println("Empieza conexión a base")

	dbDestino := "root:@tcp(127.0.0.1:3306)/cbpeh"

	//defer dbConexion.Close()
	const (
		USE_MYMYSQL = false // En caso de no funcionar un driver utiliza otro de mysql :3
	)

	driver := ""
	connstr := ""
	if USE_MYMYSQL {
		driver = "mymysql"
		connstr = dbDestino
		defer Recuperacion("producción")
	} else {
		driver = "mysql"
		connstr = dbDestino
		defer Recuperacion("Produccion")
	}

	db, err := sql.Open(driver, connstr)
	defer Recuperacion(dbDestino)
	if err != nil {
		panic(err.Error())
		log.Print("No se conecto a la base")
	}
	//dbConexion, err := sql.Open("mysql", dbDestino)
	fmt.Println("Conectado a la base destino")

	return db, nil
}

func main() {
	db, err := conexion()
	// Ahora vemos si tenemos conexión
	err = db.Ping()
	if err != nil {
		fmt.Printf("Error conectando: %v", err)
		return
	}
	// Listo, aquí ya podemos usar a db!
	fmt.Println("Conectado correctamente")

	path := "among.png"
	content, err := os.ReadFile("files/" + path)
	if err != nil {
		log.Fatal(err)
	}
	// Convert []byte to string and print to screen
	text := string(content)
	//fmt.Println(text)
	id_expediente := "CBPEH-128-2021"
	IdTipoArchivo := 6
	t := time.Now()
	//fmt.Println(t.String())
	fmt.Printf("Fecha hoy: ")
	hoy := t.Format("2006-01-02")
	tiempo := t.Format("15:04:05")
	descripcion := "IMAGEN DE COLABORACIONES RELACIONADAS AL EXPEDIENTE"
	id_perfil_carga := 2
	id_busqueda := 0
	estatus := 1

	c := Datos{
		id_expediente:   id_expediente,
		IdTipoArchivo:   IdTipoArchivo,
		hoy:             hoy,
		tiempo:          tiempo,
		usuarioCarga:    "atencion_neixar",
		path:            path,
		text:            text,
		descripcion:     descripcion,
		id_perfil_carga: id_perfil_carga,
		id_busqueda:     id_busqueda,
		estatus:         estatus,
	}

	err = insertar(c)
	if err != nil {
		fmt.Printf("Error insertando: %v", err)
	} else {
		fmt.Println("Insertado correctamente")
	}

	defer elapsed("Robot termino:")()
	time.Sleep(time.Second * 1)
	fmt.Println("Finalizando aplicación....")
}

//termina main

func insertar(c Datos) (e error) {
	db, err := conexion()
	if err != nil {
		return err
	}
	defer db.Close()

	// Preparamos para prevenir inyecciones SQL
	sentenciaPreparada, err := db.Prepare("INSERT INTO archivo_expediente (id_expediente, id_tipo_archivo, fecha_carga, hora_carga, id_usuario_carga, nombre_archivo, archivo, archivo_descripcion, id_perfil_carga, id_busqueda, estatus_expediente) VALUES(?, ?, ?,?, ?, ?,?, ?, ?,?, ?)")
	if err != nil {
		return err
	}
	defer sentenciaPreparada.Close()
	// Ejecutar sentencia, un valor por cada '?'
	_, err = sentenciaPreparada.Exec(c.id_expediente, c.IdTipoArchivo, c.hoy, c.tiempo, c.usuarioCarga, c.path, c.text, c.descripcion, c.id_perfil_carga, c.id_busqueda, c.estatus)
	if err != nil {
		return err
	}
	return nil
}

func elapsed(what string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s Finalizo en: %v\n", what, time.Since(start))
	}
}

func Recuperacion(IP string) {
	recuperado := recover()
	if recuperado != nil {
		fmt.Println("Recuperación de: ", IP, recuperado)
	}
}
