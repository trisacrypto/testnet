FROM node:14.15.0-alpine3.12

WORKDIR /

ENV PATH /node_modules/.bin:$PATH

RUN npm install --silent

COPY . ./

CMD ["npm", "start"]
