package main

import (
	"fmt"
	"log"
)

// A small demo that exercises the store and prints the result.
func main() {
	s := NewStore()

	notebook := s.Create(Item{Name: "Notebook", Price: 3.50})
	s.Create(Item{Name: "Pen", Price: 1.25})
	eraser := s.Create(Item{Name: "Eraser", Price: 0.75})

	if _, err := s.Update(notebook.ID, Item{Name: "Notebook (large)", Price: 4.75}); err != nil {
		log.Fatal(err)
	}
	if err := s.Delete(eraser.ID); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Items in store:")
	for _, it := range s.List() {
		fmt.Printf("  #%d  %-18s $%.2f\n", it.ID, it.Name, it.Price)
	}
}
