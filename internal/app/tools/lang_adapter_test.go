package tools

import "testing"

func TestCFunctionExtract(t *testing.T) {
    src := "int add(int a,int b){\nreturn a+b;\n}\n"
    got := CAdapter{}.ExtractFunction(src)
    if len(got) == 0 { t.Fatalf("empty function extract") }
}

func TestJavaClassExtract(t *testing.T) {
    src := "public class A{\npublic void x(){}\n}"
    got := JavaAdapter{}.ExtractClass(src)
    if len(got) == 0 { t.Fatalf("empty class extract") }
}

