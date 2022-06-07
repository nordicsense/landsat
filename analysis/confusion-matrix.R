nms <- c("Cloud",
    "Clean water",
    "Water with sediments",
    "Non-vegetated / Stone tundra",
    "Deciduous w/ birch, willow, grass",
    "Burnt areas, mostly new",
    "Deciduous, mostly recovered w/ birch",
    "Coniferous: 40-60% damaged or pine",
    "Coniferous, mostly spruce",
    "Wetland, mostly open",
    "Dwarf shrub, shrub, lichen")

cm <- rbind(
    c(1995,0,0,0,0,0,0,0,0,0,5),
    c(0,1908,0,1,38,0,4,15,17,9,8),
    c(0,0,239,26,2,1,2,0,0,0,21),
    c(0,22,8,1429,107,20,50,78,14,24,248),
    c(0,30,0,32,1222,5,70,182,75,168,216),
    c(0,56,0,15,20,178,2,35,25,10,11),
    c(0,0,0,22,108,0,729,57,26,5,34),
    c(0,55,0,4,70,2,19,1641,134,41,34),
    c(0,1,0,15,74,2,30,236,511,43,94),
    c(0,57,0,6,398,1,16,130,50,644,118),
    c(0,6,0,22,181,1,64,32,29,54,1611)
)



colnames(cm) <- nms
rownames(cm) <- nms

cm <- cm[11:1,,drop=FALSE]

library(RColorBrewer)
coul <- colorRampPalette(brewer.pal(8, "Blues"))(25)

pdf("confusion-matrix.pdf", width=7, height=7)

heatmap(cm, Colv = NA, Rowv = NA, col=coul, margins=c(20,20), scale="row", symm=TRUE,
    xlab="Predicted", ylab="Actual")

dev.off()
