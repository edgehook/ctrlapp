# Topic消息定义
## Server 切换 
### IoTEdge Send
**Topic**:  edge/control/${MAC}/config/msgID 
**msg**: messageContent

 	#这个内容为AppHub Server端二维码的内容
 	{
 		"url":"http://api-dccs-ensaas.hz.wise-paas.com.cn/v1/serviceCredentials/",
 		"services":[
 			{
 			"name":"mqtt",
 			"key":"082990c83844a45327cb32311f2fddd9",
 			"uid":"gangqiang.sun@advantech.com.cn",
 			"sid":"56781"
 			}
 		],
 		"bapi":"https://portal-apphub-manager-apphub-eks001.hz.wise-paas.com.cn/api/configmgr/blobkey",
 		"uname":"gangqiang.sun@advantech.com.cn"
 	}
    #如果为switch backoff 动作，则AppHub Server端二维码的内容
	{
 	}
	推荐返回控json 串，或者至少不能包含"url" 和"services" items.
    
### CtlApp Respond
**Topic**:  device/response/${MAC}/$(parent msgID) 
**msg**: status

 	{
 		"errorcode":"0",
 		"reason":""
 		"parameter":""
 	}

## 更改DeviceName (通过IotHub修改设备名) 
### IoTEdge Send
**Topic**:  edge/control/${MAC}/changename/msgID
**msg**: messageContent

	{
		"devicename":"xxx"
	}

### CtlApp Respond
**Topic**:  device/response/${MAC}/$(parent msgID
**msg**: status

	{
		"errorcode":"0",
		"reason":""
		"parameter":""		
	}

## 更改DeviceName (通过AppHub修改设备名)
###  CtlApp Send
**Topic**:  device/report/${MAC}/namechanged/$(msgID)
**msg**: messageContent	

	{
		"devicename":"xxx"
	}

### IoTEdge Respond
**Topic**:  edge/response/${MAC}/$(parent msgID)
**msg**: status	

	{
		"errorcode":"0",
		"reason":""
		"parameter":""
	}

## 设置UUID 
### IoTEdge Send
**Topic**:  edge/control/${MAC}/setuuid/msgID
**msg**: messageContent

	{
		"uuid":"xxx"
	}

### CtlApp Respond
**Topic**:  device/response/${MAC}/$(parent msgID)
**msg**: status

	{
		"errorcode":"0",
		"reason":""
	}


## 获取UUID 
### IoTEdge Send
**Topic**:  edge/control/${MAC}/getuuid/msgID
**msg**: messageContent

	{
	}

### CtlApp Respond
**Topic**:  device/response/${MAC}/$(parent msgID)
**msg**: status	

	{
		"errorcode":"0",
		"reason":""
		"parameter":{
			"uuid":"xxx"
		}
	}



# 其他说明
1. 所有通信都为同步通信
2. 所有的消息体都应该是JSON,无论回复的还是发送的消息
3. 所有的消息体都应该加密, 加密和解密方法再做讨论
4. respond消息体中的errorcode目前定义以下两种值
  - **0**: 成功 
  - **非零**: 失败, 失败可以在reason字段中说明原因

