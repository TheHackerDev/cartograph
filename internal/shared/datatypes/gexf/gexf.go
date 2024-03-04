/*******************************************************************************
 * Copyright 2018-2024 Aaron Hnatiw
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 ******************************************************************************/

package gexf

import (
	"encoding/xml"
	"fmt"
	"io"
	"time"
)

type Gexf struct {
	XMLName   xml.Name `xml:"gexf"`
	Xmlns     string   `xml:"xmlns,attr"`
	Xsi       string   `xml:"xmlns:xsi,attr"`
	SchemaLoc string   `xml:"xsi:schemaLocation,attr"`
	Version   string   `xml:"version,attr"`
	Meta      Meta
	Graph     Graph
}

type Meta struct {
	XMLName          xml.Name `xml:"meta"`
	LastModifiedDate string   `xml:"lastmodifieddate,attr"`
	Creator          string   `xml:"creator"`
	Description      string   `xml:"description"`
	Keywords         string   `xml:"keywords"`
}

type Graph struct {
	XMLName    xml.Name `xml:"graph"`
	Mode       string   `xml:"mode,attr"`
	TimeFormat string   `xml:"timeformat,attr,omitempty"`
	Attributes Attributes
	Nodes      Nodes
	Edges      Edges
}

type Attributes struct {
	XMLName    xml.Name    `xml:"attributes"`
	Attributes []Attribute `xml:"attribute"`
	Class      string      `xml:"class,attr"`
}

type Attribute struct {
	XMLName xml.Name `xml:"attribute"`
	Id      string   `xml:"id,attr"`
	Title   string   `xml:"title,attr"`
	Type    string   `xml:"type,attr"`
}

type Nodes struct {
	XMLName xml.Name `xml:"nodes"`
	Nodes   []Node   `xml:"node"`
}

type Node struct {
	XMLName   xml.Name  `xml:"node"`
	Id        string    `xml:"id,attr"`
	Pid       string    `xml:"pid,attr,omitempty"`
	Label     string    `xml:"label,attr"`
	Start     string    `xml:"start,attr,omitempty"`
	End       string    `xml:"end,attr,omitempty"`
	Attvalues Attvalues `xml:"attvalues"`
}

type Attvalues struct {
	XMLName   xml.Name   `xml:"attvalues"`
	Attvalues []Attvalue `xml:"attvalue"`
}

type Attvalue struct {
	XMLName xml.Name `xml:"attvalue"`
	For     string   `xml:"for,attr"`
	Value   string   `xml:"value,attr"`
}

type Edges struct {
	XMLName xml.Name `xml:"edges"`
	Edges   []Edge   `xml:"edge"`
}

type Edge struct {
	XMLName xml.Name `xml:"edge"`
	Id      string   `xml:"id,attr,omitempty"`
	Source  string   `xml:"source,attr"`
	Target  string   `xml:"target,attr"`
	Start   string   `xml:"start,attr,omitempty"`
	End     string   `xml:"end,attr,omitempty"`
}

func (g *Gexf) CreateXML(w io.Writer, description, keywords string) error {
	// Fill out the Gexf struct with default values
	g.Xmlns = "http://gexf.net/1.3"
	g.Xsi = "http://www.w3.org/2001/XMLSchema-instance"
	g.SchemaLoc = "http://gexf.net/1.3 http://gexf.net/1.3/gexf.xsd"
	g.Version = "1.3"

	// Fill out the default values in the Meta struct
	g.Meta.LastModifiedDate = time.Now().Format(time.DateOnly)
	g.Meta.Creator = "The Hacker Dev"

	// Add the provided description and keywords to the Meta struct
	g.Meta.Description = description
	g.Meta.Keywords = keywords

	// Create the XML encoder
	e := xml.NewEncoder(w)

	// Start with the XML processing instructions
	if err := e.EncodeToken(xml.ProcInst{"xml", []byte(`version="1.0" encoding="UTF-8"`)}); err != nil {
		return fmt.Errorf("unable to encode XML processing instructions: %w", err)
	}

	// Encode the rest of the Gexf struct
	if err := e.Encode(g); err != nil {
		return fmt.Errorf("unable to encode Gexf struct: %w", err)
	}

	return nil
}
