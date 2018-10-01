//hover-overlay 


document.onmousemove = function(){
    // Get browser width
    var w = Math.max(document.documentElement.clientWidth, window.innerWidth || 0);
    var scrollHero = document.querySelector(".code-intellify-container").scrollLeft;
    var scrollModal = document.querySelector(".code-intellify-container-modal").scrollLeft;
    console.log(scrollModal)
    // Get posistion of hover

    var tooltip = document.querySelector(".hover-overlay")
    var rect = tooltip.getBoundingClientRect();

    if (document.querySelector(".code-intellify-container-modal").classList.contains("modal-open")){
            
            var left = Number(tooltip.style.left.replace(/px/g, ''))
    
            //Get ToolTip Width
            tW = Number(rect.right - rect.left)
        if ( rect.right >= w ) {
            newLeft = left - tW + 100;
            tooltip.style.transform = "translateX("+scrollModal+"px)";
            tooltip.style.left = newLeft+"px";
        } else {
            // tooltip.style.left = (scrollModal - left)+"px";
            tooltip.style.left = left+"px";
            tooltip.style.transform = "translateX(-"+scrollModal+"px)";

        }
    } else {
            
            var left = Number(tooltip.style.left.replace(/px/g, ''))
    
            //Get ToolTip Width
            tW = Number(rect.right - rect.left)
        if ( rect.right >= w ) {
            console.log("it's outside – adjust");
            tooltip.style.transform = "translateX("+scrollHero+"px)";
            tooltip.style.left = newLeft+"px";
    
        } else {
            // tooltip.style.left = (scrollHero - left)+"px";
            tooltip.style.left = left+"px";
            tooltip.style.transform = "translateX(-"+scrollHero+"px)";
        }
    }
   

  }