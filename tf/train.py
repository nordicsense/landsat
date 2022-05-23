import pandas as pd
import tensorflow as tf

df = pd.read_csv('/Users/osklyar/Data/Landsat/analysis/model/trainingdata.csv')
x = df.drop(['clazz', 'clazzid'], axis=1)
y = df['clazzid']

df = pd.read_csv('/Users/osklyar/Data/Landsat/analysis/model/trainingdata-test.csv')
x_test = df.drop(['clazz', 'clazzid'], axis=1)
y_test = df['clazzid']

tf.random.set_seed(42)

model = tf.keras.Sequential([
    tf.keras.layers.Dense(512, activation='relu', name="outer"),
    tf.keras.layers.Dense(512, activation='relu', name="inner"),
    tf.keras.layers.Dense(13, activation='softmax', name="clazzifier"), # 13 classes
], name="sequential")

model.compile(
    optimizer=tf.keras.optimizers.Adam(learning_rate=0.01),
    # optimizer=tf.keras.optimizers.SGD(lr=0.01, decay=1e-6, momentum=0.9, nesterov=True),
    loss=tf.keras.losses.SparseCategoricalCrossentropy(),
    metrics=['accuracy']
)

epochs = model.fit(x, y, epochs=10)

model.evaluate(x_test, y_test, verbose=2)

# saved_model_cli show --dir tf.model --all
model.save("/Users/osklyar/Data/Landsat/analysis/model/tf.model")

