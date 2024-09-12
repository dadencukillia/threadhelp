FROM node:22-alpine3.19
WORKDIR /home/work
COPY ./package*.json ./
RUN npm i
COPY . .
CMD npm run build
