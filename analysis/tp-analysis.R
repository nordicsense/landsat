# Temperature & Precipitation analysis

t <- read.csv("~/Downloads/temperature.csv")

t <- do.call("rbind", apply(t, 1, function(x) {
    data.frame(date=as.Date(sprintf("%d-%d-01", x[1], 1:12)), temp=x[2:13])
}))





t <- read.csv("~/github.com/nordicsense/landsat/analysis/temperature.csv")
p <- read.csv("~/github.com/nordicsense/landsat/analysis/precipitation.csv")

xx <- t$temp
yy <- p$precip

x <- xx[5:(length(xx)-5)]

y1 <- yy[1:(length(yy)-9)]
y2 <- yy[2:(length(yy)-8)]
y3 <- yy[3:(length(yy)-7)]
y4 <- yy[4:(length(yy)-6)]
y5 <- yy[5:(length(yy)-5)]
y6 <- yy[6:(length(yy)-4)]
y7 <- yy[7:(length(yy)-3)]
y8 <- yy[8:(length(yy)-2)]
y9 <- yy[9:(length(yy)-1)]
y10 <- yy[10:length(yy)]

sapply(list(y1, y2, y3, y4, y5, y6, y7, y8, y9, y10), function(y) {
    cor(x, y, use="pairwise.complete.obs", method="pearson")
})

plot(as.Date(t$date), t$temp/p$precip, type="l", col="red", ylim=c(-5,5))

dt <- as.Date(t$date[6:(nrow(t)-6)])
tt <- t$temp[6:(nrow(t)-6)]
pp <- p$precip[1:(nrow(p)-11)]

pdf("~/Downloads/temp.pdf", width=11, height=8)
par(mfrow=c(4,1), mar=c(4,4,1,1))

plot(dt, tt/p$precip[6:(nrow(p)-6)], type="l", col="black", ylim=c(-4,4), xaxt="n", xlim=as.Date(c("1982-01-01","1992-01-01")), xlab="", ylab="temp/precip")
axis.Date(1,at=dt,labels=format(dt,"%y/%m"),las=2)
#lines(dt, tt/p$precip[3:(nrow(p)-9)], col="blue")
lines(dt, tt/p$precip[1:(nrow(p)-11)], col="red")
legend("topright", legend=c("same time", "precip 6m earlier"), col=c("black", "red"), lwd=2)


plot(dt, tt/p$precip[6:(nrow(p)-6)], type="l", col="black", ylim=c(-4,4), xaxt="n", xlim=as.Date(c("1992-01-01","2002-01-01")), xlab="", ylab="temp/precip")
axis.Date(1,at=dt,labels=format(dt,"%y/%m"),las=2)
#lines(dt, tt/p$precip[3:(nrow(p)-9)], col="blue")
lines(dt, tt/p$precip[1:(nrow(p)-11)], col="red")

plot(dt, tt/p$precip[6:(nrow(p)-6)], type="l", col="black", ylim=c(-4,4), xaxt="n", xlim=as.Date(c("2002-01-01","2012-01-01")), xlab="", ylab="temp/precip")
axis.Date(1,at=dt,labels=format(dt,"%y/%m"),las=2)
#lines(dt, tt/p$precip[3:(nrow(p)-9)], col="blue")
lines(dt, tt/p$precip[1:(nrow(p)-11)], col="red")

plot(dt, tt/p$precip[6:(nrow(p)-6)], type="l", col="black", ylim=c(-4,4), xaxt="n", xlim=as.Date(c("2012-01-01","2022-01-01")), xlab="", ylab="temp/precip")
axis.Date(1,at=dt,labels=format(dt,"%y/%m"),las=2)
#lines(dt, tt/p$precip[3:(nrow(p)-9)], col="blue")
lines(dt, tt/p$precip[1:(nrow(p)-11)], col="red")

dev.off()
