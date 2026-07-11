package gedcom

type Header struct {
}

func (h *Header) String() string {
	b := NewBlock()
	b.Add("HEAD")
	b.AddDown("CHAR UTF-8").Up()
	b.AddDown("GEDC").AddDown("VERS 5.5.5")
	b.Add("FORM LINEAGE-LINKED")
	return b.String()
}

// StahPete1762
// StahPete1762StahGeno1762
// StahGeno1762
