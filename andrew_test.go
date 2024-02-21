package andrew_test

// import (
// 	"github.com/playtechnique/andrew"
// 	"os"
// 	"slices"
// 	"testing"
// )
//
// func TestRetrieveEmptySetWhenOnlyIndexHtml(t *testing.T) {
// 	testDir := t.TempDir()
//
// 	contentRoot := testDir + "/onlyIndexHtml"
// 	os.Mkdir(contentRoot, os.ModePerm)
// 	os.WriteFile(contentRoot+"/index.html", []byte{}, os.ModePerm)
//
// 	expected := []string{}
// 	actual, err := andrew.GetLinks(contentRoot)
//
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	if !slices.Equal(expected, actual) {
// 		t.Fatalf("expected to generate %s actually received %s", expected, actual)
// 	}
//
// }
//
// func TestGenerateLinkToOneFile(t *testing.T) {
// 	testDir := t.TempDir()
//
// 	contentRoot := "website"
// 	absPath := testDir + "/" + contentRoot
// 	err := os.Mkdir(absPath, os.ModePerm)
//
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	os.WriteFile(absPath+"/index.html", []byte{}, os.ModePerm)
// 	os.WriteFile(absPath+"/somearticle.html", []byte{}, os.ModePerm)
// 	os.WriteFile(absPath+"/main.css", []byte{}, os.ModePerm)
// 	os.WriteFile(absPath+"/main.js", []byte{}, os.ModePerm)
//
// 	expected := []string{"<a href=somearticle.html>somearticle</a>"}
//
// 	err = os.Chdir(testDir)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	actual, err := andrew.GetLinks(contentRoot)
//
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	if !slices.Equal(expected, actual) {
// 		t.Fatalf("expected to generate %s actually received %s", expected, actual)
// 	}
//
// }
