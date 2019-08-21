package generator

type Generator interface {
	Generate() (string, error)
}
