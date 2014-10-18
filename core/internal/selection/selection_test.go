package selection_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti/core/internal/mocks"
	. "github.com/sclevine/agouti/core/internal/selection"
	"github.com/sclevine/agouti/core/internal/types"
)

var _ = Describe("Selection", func() {
	var (
		selection types.Selection
		client    *mocks.Client
		element   *mocks.Element
	)

	BeforeEach(func() {
		client = &mocks.Client{}
		element = &mocks.Element{}
		selection = &Selection{Client: client}
		selection = selection.Find("#selector")
	})

	ItShouldEnsureASingleElement := func(matcher func() error) {
		Context("ensures a single element is returned", func() {
			It("should return an error with the number of elements", func() {
				client.GetElementsCall.ReturnElements = []types.Element{element, element}
				Expect(matcher()).To(MatchError("failed to retrieve element with 'CSS: #selector': multiple elements (2) were selected"))
			})
		})
	}

	Describe("most methods: retrieving elements", func() {
		var (
			parentOne *mocks.Element
			parentTwo *mocks.Element
			count     int
		)

		BeforeEach(func() {
			parentOne = &mocks.Element{}
			parentTwo = &mocks.Element{}
			parentOne.GetElementsCall.ReturnElements = []types.Element{&mocks.Element{}, &mocks.Element{}}
			parentTwo.GetElementsCall.ReturnElements = []types.Element{&mocks.Element{}, &mocks.Element{}}
			client.GetElementsCall.ReturnElements = []types.Element{parentOne, parentTwo}
		})

		Context("when successful without indices", func() {
			BeforeEach(func() {
				selection = selection.FindXPath("children")
				count, _ = selection.Count()
			})

			It("should retrieve the parent elements using the client", func() {
				Expect(client.GetElementsCall.Selector).To(Equal(types.Selector{Using: "css selector", Value: "#selector"}))
			})

			It("should retrieve the child elements of the parent selector", func() {
				Expect(parentOne.GetElementsCall.Selector).To(Equal(types.Selector{Using: "xpath", Value: "children"}))
				Expect(parentTwo.GetElementsCall.Selector).To(Equal(types.Selector{Using: "xpath", Value: "children"}))
			})

			It("should return all child elements of the terminal selector", func() {
				Expect(count).To(Equal(4))
			})
		})

		Context("when successful with indices", func() {
			BeforeEach(func() {
				selection.At(1).FindXPath("children").At(1).Click()
			})

			It("should retrieve the parent elements using the client", func() {
				Expect(client.GetElementsCall.Selector).To(Equal(types.Selector{Using: "css selector", Value: "#selector", Index: 1, Indexed: true}))
			})

			It("should retrieve the child elements of the parent selector", func() {
				Expect(parentOne.GetElementsCall.Selector.Using).To(BeEmpty())
				Expect(parentTwo.GetElementsCall.Selector).To(Equal(types.Selector{Using: "xpath", Value: "children", Index: 1, Indexed: true}))
			})

			It("should return all child elements of the terminal selector", func() {
				clickedElement := parentTwo.GetElementsCall.ReturnElements[1].(*mocks.Element)
				Expect(clickedElement.ClickCall.Called).To(BeTrue())
			})
		})

		Context("when there is no selection", func() {
			BeforeEach(func() {
				selection = &Selection{Client: client}
			})

			It("should return an error", func() {
				_, err := selection.Count()
				Expect(err).To(MatchError("failed to retrieve elements for '': empty selection"))
			})
		})

		Context("when retrieving the parent elements fails", func() {
			BeforeEach(func() {
				selection = selection.FindXPath("children")
				client.GetElementsCall.Err = errors.New("some error")
			})

			It("should return the error", func() {
				_, err := selection.Count()
				Expect(err).To(MatchError("failed to retrieve elements for 'CSS: #selector | XPath: children': some error"))
			})
		})

		Context("when retrieving any of the child elements fails", func() {
			BeforeEach(func() {
				selection = selection.FindXPath("children")
				parentTwo.GetElementsCall.Err = errors.New("some error")
			})

			It("should return the error", func() {
				_, err := selection.Count()
				Expect(err).To(MatchError("failed to retrieve elements for 'CSS: #selector | XPath: children': some error"))
			})
		})

		Context("when the first selection index is out of range", func() {
			It("should return an error with the index and total number of elements", func() {
				Expect(selection.At(2).Click()).To(MatchError("failed to retrieve element with 'CSS: #selector [2]': element index out of range (>1)"))
			})
		})

		Context("when subsequent selection indices are out of range", func() {
			It("should return an error with the index and total number of elements", func() {
				Expect(selection.At(0).Find("#selector").At(2).Click()).To(MatchError("failed to retrieve element with 'CSS: #selector [0] | CSS: #selector [2]': element index out of range (>1)"))
			})
		})
	})

	Describe("#At & most methods: retrieving the selected element", func() {
		It("should request an element from the client using the element's selector", func() {
			selection.Click()
			Expect(client.GetElementsCall.Selector).To(Equal(types.Selector{Using: "css selector", Value: "#selector"}))
		})

		Context("when the client fails to retrieve any elements", func() {
			It("should return error from the client", func() {
				client.GetElementsCall.Err = errors.New("some error")
				Expect(selection.Click()).To(MatchError("failed to retrieve element with 'CSS: #selector': some error"))
			})
		})

		Context("when the client retrieves zero elements", func() {
			It("should fail with an error indicating there were no elements", func() {
				client.GetElementsCall.ReturnElements = []types.Element{}
				Expect(selection.Click()).To(MatchError("failed to retrieve element with 'CSS: #selector': no element found"))
			})
		})

		Context("when the client retrieves more than one element and indexing is disabled", func() {
			It("should return an error with the number of elements", func() {
				client.GetElementsCall.ReturnElements = []types.Element{element, element}
				Expect(selection.Click()).To(MatchError("failed to retrieve element with 'CSS: #selector': multiple elements (2) were selected"))
			})
		})
	})

	Describe("#Find", func() {
		Context("when there is no selection", func() {
			It("should add a new css selector to the selection", func() {
				selection := &Selection{Client: client}
				Expect(selection.Find("#selector").String()).To(Equal("CSS: #selector"))
			})
		})

		Context("when the selection ends with an xpath selector", func() {
			It("should add a new css selector to the selection", func() {
				xpath := selection.FindXPath("//subselector")
				Expect(xpath.Find("#subselector").String()).To(Equal("CSS: #selector | XPath: //subselector | CSS: #subselector"))
			})
		})

		Context("when the selection ends with an unindexed CSS selector", func() {
			It("should modifie the terminal css selector to include the new selector", func() {
				Expect(selection.Find("#subselector").String()).To(Equal("CSS: #selector #subselector"))
			})
		})

		Context("when the selection ends with an indexed CSS selector", func() {
			It("should add a new css selector to the selection", func() {
				Expect(selection.At(0).Find("#subselector").String()).To(Equal("CSS: #selector [0] | CSS: #subselector"))
			})
		})

		Context("when two CSS selections are created from the same XPath parent", func() {
			It("should not overwrite the first created child", func() {
				selection := &Selection{Client: client}
				parent := selection.FindXPath("//one").FindXPath("//two").FindXPath("//parent")
				firstChild := parent.Find("#firstChild")
				parent.Find("#secondChild")
				Expect(firstChild.String()).To(Equal("XPath: //one | XPath: //two | XPath: //parent | CSS: #firstChild"))
			})
		})
	})

	Describe("#FindXPath", func() {
		It("should add a new XPath selector to the selection", func() {
			Expect(selection.FindXPath("//subselector").String()).To(Equal("CSS: #selector | XPath: //subselector"))
		})
	})

	Describe("#FindLink", func() {
		It("should add a new 'link text' selector to the selection", func() {
			Expect(selection.FindLink("some text").String()).To(Equal(`CSS: #selector | Link: "some text"`))
		})
	})

	Describe("#FindByLabel", func() {
		It("should add an XPath selector for finding by label", func() {
			Expect(selection.FindByLabel("label name").String()).To(Equal(`CSS: #selector | XPath: //input[@id=(//label[normalize-space(text())="label name"]/@for)] | //label[normalize-space(text())="label name"]/input`))
		})
	})

	Describe("#All", func() {
		It("should return a MultiSelection created from the Selection", func() {
			Expect(selection.All().String()).To(Equal(`CSS: #selector - All`))
		})
	})

	Describe("#String", func() {
		It("should return the separated selectors", func() {
			Expect(selection.FindXPath("//subselector").String()).To(Equal("CSS: #selector | XPath: //subselector"))
		})

		Context("when indexed via At(index)", func() {
			It("should append [index] to the indexed selectors", func() {
				Expect(selection.At(2).FindXPath("//subselector").At(1).String()).To(Equal("CSS: #selector [2] | XPath: //subselector [1]"))
			})
		})
	})

	Describe("#Count", func() {
		BeforeEach(func() {
			client.GetElementsCall.ReturnElements = []types.Element{element, element}
		})

		It("should request elements from the client using the provided selector", func() {
			selection.Count()
			Expect(client.GetElementsCall.Selector).To(Equal(types.Selector{Using: "css selector", Value: "#selector"}))
		})

		Context("when the client succeeds in retrieving the elements", func() {
			It("should return the text", func() {
				count, _ := selection.Count()
				Expect(count).To(Equal(2))
			})

			It("should not return an error", func() {
				_, err := selection.Count()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the the client fails to retrieve the elements", func() {
			BeforeEach(func() {
				client.GetElementsCall.Err = errors.New("some error")
			})

			It("should return an error", func() {
				_, err := selection.Count()
				Expect(err).To(MatchError("failed to retrieve elements for 'CSS: #selector': some error"))
			})
		})
	})

	Describe("#EqualsElement", func() {
		var (
			otherClient    *mocks.Client
			otherSelection types.Selection
			otherElement   *mocks.Element
		)

		BeforeEach(func() {
			client.GetElementsCall.ReturnElements = []types.Element{element}
			otherClient = &mocks.Client{}
			otherSelection = &Selection{Client: otherClient}
			otherSelection = otherSelection.Find("#other_selector")
			otherElement = &mocks.Element{}
			otherClient.GetElementsCall.ReturnElements = []types.Element{otherElement}
		})

		ItShouldEnsureASingleElement(func() error {
			_, err := selection.EqualsElement(otherSelection)
			return err
		})

		It("should ensure that the other selection is a single element", func() {
			otherClient.GetElementsCall.ReturnElements = []types.Element{element, element}
			_, err := selection.EqualsElement(otherSelection)
			Expect(err).To(MatchError("failed to retrieve element with 'CSS: #other_selector': multiple elements (2) were selected"))
		})

		It("should compare the selection elements for equality", func() {
			selection.EqualsElement(otherSelection)
			Expect(element.IsEqualToCall.Element).To(Equal(otherElement))
		})

		Context("if the provided element is not a *Selection", func() {
			It("should return an error", func() {
				_, err := selection.EqualsElement("not a selection")
				Expect(err).To(MatchError("provided object is not a selection"))
			})
		})

		Context("if the client fails to compare the elements", func() {
			It("should return an error", func() {
				element.IsEqualToCall.Err = errors.New("some error")
				_, err := selection.EqualsElement(otherSelection)
				Expect(err).To(MatchError("failed to compare 'CSS: #selector' to 'CSS: #other_selector': some error"))
			})
		})

		Context("if the client succeeds in comparing the elements", func() {
			It("should return true if they are equal", func() {
				element.IsEqualToCall.ReturnEquals = true
				equal, _ := selection.EqualsElement(otherSelection)
				Expect(equal).To(BeTrue())
			})

			It("should return false if they are not equal", func() {
				element.IsEqualToCall.ReturnEquals = false
				equal, _ := selection.EqualsElement(otherSelection)
				Expect(equal).To(BeFalse())
			})

			It("should not return an error", func() {
				_, err := selection.EqualsElement(otherSelection)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
