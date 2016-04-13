class CoffeeBorker
  bork: () ->
    throw new Error("Bork from coffeescript")
    
window.coffee_bork = () -> new CoffeeBorker().bork()
