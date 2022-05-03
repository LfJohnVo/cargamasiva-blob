package main

// It's importing the packages that are needed for the program to run.
import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/schollz/progressbar/v3"
)

// It's a struct with a bunch of fields.
// @property {string} id_expediente - The id of the file
// @property {int} IdTipoArchivo - This is the type of file that is being uploaded.
// @property {string} hoy - today's date
// @property {string} tiempo - time of the day
// @property {string} usuarioCarga - The user who uploaded the file
// @property {string} path - The path to the file
// @property {string} text - The text to be searched for.
// @property {string} descripcion - Description of the file
// @property {int} id_perfil_carga - This is the id of the profile that is uploading the file.
// @property {int} id_busqueda - This is the id of the search.
// @property {int} estatus - 1 = not found, 2 = found, 3 = found and processed
// @property {int} estatus_localizado - 0 = not found, 1 = found
type Datos struct {
	id_expediente      string
	IdTipoArchivo      int
	hoy                string
	tiempo             string
	usuarioCarga       string
	path               string
	text               string
	descripcion        string
	id_perfil_carga    int
	id_busqueda        int
	estatus            int
	estatus_localizado int
}

// It opens a connection to a MySQL database, and returns a pointer to a sql.DB object
func conexion() (db *sql.DB, e error) {
	dbDestino := "root:@tcp(127.0.0.1:3306)/cbpeh"

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

	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(0)
	db.SetMaxOpenConns(0)

	defer Recuperacion(dbDestino)
	if err != nil {
		panic(err.Error())
		log.Println("No se conecto a la base")
	} else {
		//fmt.Printf("Conectado a la base destino %v\n", dbDestino)
	}

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

	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)

	files, err := ioutil.ReadDir("files")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Contando ....")
	time.Sleep(time.Second * 1)
	fmt.Printf("Total archivos: ")
	fmt.Println(len(files))
	fmt.Println("EMPIEZA ROBOT")
	log.Printf("/****************** EMPIEZA Robot *******************/\n")

	bar := progressbar.Default(int64(len(files)))

	i := 0
	for _, file := range files {
		bar.Add(1)
		archivo := file.Name()

		path := archivo
		content, err := os.ReadFile("files/" + path)
		if err != nil {
			log.Fatal(err)
		}

		fileExtension := filepath.Ext("files/" + path)

		// Convert []byte to string and print to screen
		text := string(content)
		id_expediente := archivo[0:14]

		var contaExiste int
		errore := db.QueryRow("SELECT count(id_expediente) FROM cbpeh.expediente WHERE id_expediente = ?", id_expediente).Scan(&contaExiste)
		switch {
		case errore != nil:
			fmt.Println("Error al consultar primer cont")
		default:
			if contaExiste >= 1 {
				var count int
				error := db.QueryRow("SELECT count(id_expediente) FROM cbpeh.archivo_expediente WHERE id_expediente = ?", id_expediente).Scan(&count)
				switch {
				case error != nil:
					fmt.Println("Error al consultar segundo cont")
				default:
					//fmt.Printf("Cantidad de registros : %v\n", count)
					if count >= 1 {
						//Update db
						stmt, e := db.Prepare("update cbpeh.archivo_expediente set archivo = ? where id_expediente = ? ")

						if e != nil {
							log.Fatal(e)
						}
						// execute
						_, e = stmt.Exec(text, id_expediente)
						if e != nil {
							log.Fatal(e)
						}
						log.Printf("Ya existe un registro, se actualizo: %v \n", archivo)
						//continue
					} else {
						IdTipoArchivo := 0
						if fileExtension == ".doc" {
							IdTipoArchivo = 1
						} else if fileExtension == ".jfif" {
							IdTipoArchivo = 2
						} else if fileExtension == ".jpeg" {
							IdTipoArchivo = 3
						} else if fileExtension == ".jpg" {
							IdTipoArchivo = 4
						} else if fileExtension == ".pdf" {
							IdTipoArchivo = 5
						} else if fileExtension == ".png" {
							IdTipoArchivo = 6
						} else if fileExtension == ".docx" {
							IdTipoArchivo = 7
						} else {
							fmt.Println(fileExtension, "has multiple digits")
						}

						t := time.Now()
						hoy := t.Format("2006-01-02")
						tiempo := t.Format("15:04:05")
						descripcion := "CARGA HISTORICA"
						id_perfil_carga := 2
						id_busqueda := 0
						estatus := 1
						estatus_localizado := 0

						c := Datos{
							id_expediente:      id_expediente,
							IdTipoArchivo:      IdTipoArchivo,
							hoy:                hoy,
							tiempo:             tiempo,
							usuarioCarga:       "atencion_neixar",
							path:               path,
							text:               text,
							descripcion:        descripcion,
							id_perfil_carga:    id_perfil_carga,
							id_busqueda:        id_busqueda,
							estatus:            estatus,
							estatus_localizado: estatus_localizado,
						}

						err = insertar(c)
						if err != nil {
							log.Printf("Error insertando: %v %v\n", err, archivo)
						} else {
							log.Printf("Insertado correctamente: %v \n", archivo)
						}

						time.Sleep(time.Second * 1)

					}
				}

				i++
			} else {
				log.Printf("Este archivo no tiene registro en la tabla expedientes: %v %v\n", err, archivo)
			}
		}

	}

	log.Printf("/******************* Robot Termino **********************/ \n")
	defer elapsed("Robot termino:")()
	fmt.Println("Finalizando aplicación....")
	time.Sleep(time.Second * 3)

}

//termina main

func insertar(c Datos) (e error) {
	db, err := conexion()
	if err != nil {
		return err
	}
	defer db.Close()

	// Preparamos para prevenir inyecciones SQL
	sentenciaPreparada, err := db.Prepare("INSERT INTO archivo_expediente (id_expediente, id_tipo_archivo, fecha_carga, hora_carga, id_usuario_carga, nombre_archivo, archivo, archivo_descripcion, id_perfil_carga, id_busqueda, estatus_expediente, estatus_localizado) VALUES(?, ?, ?,?, ?, ?,?, ?, ?,?, ?,?)")
	if err != nil {
		return err
	}
	defer sentenciaPreparada.Close()
	// Ejecutar sentencia, un valor por cada '?'
	_, err = sentenciaPreparada.Exec(c.id_expediente, c.IdTipoArchivo, c.hoy, c.tiempo, c.usuarioCarga, c.path, c.text, c.descripcion, c.id_perfil_carga, c.id_busqueda, c.estatus, c.estatus_localizado)
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
