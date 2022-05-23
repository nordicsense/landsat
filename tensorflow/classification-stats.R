
nms <- c("cloud", "water", "impact-water", "agricultural", "burnt-new", "burnt-old", "impact-damaged",
    "impact-nonvegetated", "nonvegetated", "tundra", "wetland", "coniferous", "decidious")


data <- read.csv("~/Data/Landsat/analysis/model/trainingdata.csv")
data <- split(data, data$clazz)
data <- data[nms]


sapply(data, nrow)

pdf(file="~/Data/Landsat/analysis/model/trainingdata.pdf", height=8.4, width=11.6)
datasets <- c("band1", "band2", "band3", "band4", "band5", "band7", "ndvi", "nbr", "ndwi") # "nbr2"
par(mar=c(4,10,2,2), mfrow=c(3,3))
for (dsname in datasets) {
    boxplot(lapply(data, "[[", dsname), horizontal=TRUE, las=2, main=dsname)
}
dev.off()


cm <- array(c(
    c(500, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0),
    c(0, 472, 0, 0, 0, 2, 0, 0, 0, 0, 3, 4, 19),
    c(0, 0, 414, 0, 1, 1, 0, 16, 60, 2, 4, 0, 2),
    c(0, 1, 0, 79, 1, 1, 2, 0, 0, 39, 7, 12, 55),
    c(0, 77, 0, 1, 231, 4, 29, 15, 5, 2, 19, 60, 57),
    c(0, 0, 0, 0, 0, 343, 6, 0, 52, 1, 16, 19, 63),
    c(0, 34, 0, 0, 10, 2, 338, 6, 22, 10, 23, 20, 35),
    c(0, 2, 0, 0, 0, 8, 8, 298, 114, 17, 23, 4, 26),
    c(0, 0, 0, 0, 1, 13, 14, 22, 391, 26, 8, 9, 16),
    c(0, 1, 0, 0, 0, 4, 3, 4, 45, 333, 27, 7, 76),
    c(0, 3, 0, 2, 1, 0, 2, 2, 2, 42, 321, 50, 75),
    c(0, 13, 0, 2, 0, 7, 3, 1, 0, 10, 38, 366, 60),
    c(0, 9, 0, 3, 1, 9, 0, 0, 23, 28, 54, 17, 356)), dim=c(13,13))
colnames(cm) <- nms
rownames(cm) <- nms

library(RColorBrewer)

coul <- colorRampPalette(brewer.pal(8, "Blues"))(25)

heatmap(cm, Colv = NA, Rowv = NA, col=coul)
