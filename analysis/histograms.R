root <- "/Users/osklyar/Data/Landsat/analysis/training"
pattern <- "T1_fix_hist"

datalist <- lapply(list.files(root, pattern = paste(pattern, "csv", sep = ".")), function(fname) {
  data <- read.csv(paste(root, fname, sep = "/"))
  data <- cbind(image = substr(fname, 1, 25), data)
  colnames(data) <- c("image", "band", "pos", "x", "f")
  data
})

data <- do.call("rbind", datalist)

images <- sort(unique(data$image))
indexes <- 1:length(images)

overrideXlim <- TRUE

pdf(paste(root, paste(pattern, "pdf", sep = "."), sep = "/"), width = 11, height = 7)
na <- lapply(split(data, data$band), function(x) {
  band <- unique(x$band)
  xlim <- range(x$x)
  if (overrideXlim) {
    if (band == 6) {
      xlim <- c(0, 2000)
    } else {
      xlim <- c(0, 0.5)
    }
  }
  plot(xlim, c(0, 0.3), type = "n", main = paste("band", band))
  for (i in indexes) {
    y <- x[x$image == images[i], , drop = FALSE]
    lines(y$x, y$f / sum(y$f), lwd = 2, col = i, lty = i)
  }
  legend("topright", legend = images, lwd = 2, col = indexes, lty = indexes)
  if (i < length(images)) {
    plot.new()
  }
})
dev.off()


plot(c(0., 0.15), c(0, 1e7))
for (i in 1:length(images)) {
  image <- images[i]
  xx <- data[data$image == image & data$band == 7,]
  lines(xx$x, xx$f, col = i)
}
legend("topright", legend = images, lwd = 2, col = 1:length(images))
