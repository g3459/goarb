library CRouter{

    bool internal constant FRP=true;
    
    uint internal constant PID_MASK=0xff000000000000000000000000000000000000000000000000000000;
    uint internal constant UNIV2_PID=0x01000000000000000000000000000000000000000000000000000000;
    uint internal constant UNIV3_PID=0;
    uint internal constant ALGB_PID=0x02000000000000000000000000000000000000000000000000000000;

    function findRoutes(uint8 maxLen,uint8 t,uint amIn,bytes[][] memory pools) public view returns (uint[] memory amounts,bytes[] memory calls){
        unchecked{
            return findRoutesInt(maxLen, t, amIn, pools);
        }
    }

    function findRoutesInt(uint8 maxLen,uint8 t,uint amIn,bytes[][] memory pools) internal view returns (uint[] memory amounts,bytes[] memory calls){
        unchecked{
            amounts=new uint[](pools.length);
            amounts[t]=amIn;
            calls=new bytes[](pools.length);
            findRoutes(maxLen*0x20,pools,amounts,calls);
        }
    }

    function findRoutes(uint8 maxLen,bytes[][] memory pools,uint[] memory amounts,bytes[] memory calls) internal view{
        unchecked{
            uint[] memory gasFees=new uint[](pools.length);
            uint updated=type(uint).max>>(256-pools.length);
            while (updated!=0){
                for (uint t0; t0 < pools.length; t0++){
                    if(updated&(1<<t0)==0){
                        continue;
                    }
                    updated^=1<<t0;
                    if(amounts[t0]==0 || calls[t0].length==maxLen){
                        continue;
                    }
                    for (uint t1; t1 < pools.length; t1++){
                        if(t0==t1){
                            continue;
                        }
                        bytes memory _pools;
                        bool direc;
                        if(pools[t0].length>0 && pools[t0][t1].length>0){
                            direc=true;
                            _pools=pools[t0][t1];
                        }else if(pools[t1].length>0 && pools[t1][t0].length>0){
                            _pools=pools[t1][t0];
                        }else{
                            continue;
                        }
                        uint eth = t1==0?0:amounts[0];
                        (uint hAmOut,uint poolCall) = quotePools(amounts[t0]-1,eth,direc,_pools);
                        uint gasNew = gasFees[t0]+callGas(poolCall);
                        {
                            uint gasFeeNew = gasNew * tx.gasprice;
                            uint gasFeeCurrent = gasFees[t1] * tx.gasprice;
                            if(eth!=0){
                                gasFeeNew=(hAmOut*gasFeeNew)/eth;
                                gasFeeCurrent=(hAmOut*gasFeeCurrent)/eth;
                            }
                            if(hAmOut<=gasFeeNew){
                                continue;
                            }
                            if(int(hAmOut-gasFeeNew)<=int(amounts[t1]-gasFeeCurrent)){
                                continue;
                            }
                        }
                        if(poolInCalls(calls[t0],uint160(poolCall))){
                            continue;
                        }
                        amounts[t1]=hAmOut-1;
                        gasFees[t1]=gasNew;
                        uint amIn56bit=compress56bit(amounts[t0]);
                        poolCall=(poolCall&0x7fffffffff00000000000000ffffffffffffffffffffffffffffffffffffffff)|(amIn56bit<<160);
                        if(direc) poolCall|=0x8000000000000000000000000000000000000000000000000000000000000000;
                        calls[t1]=bytes.concat(calls[t0],abi.encode(poolCall));
                        updated|=1<<t1;
                    }
                }
            }
        }
    }

    // function decompress56bit(uint compressed)internal pure returns (uint){
    //     unchecked{
    //         return uint(uint48(compressed>>8))<<uint8(compressed);
    //     }
    // }

    function quotePools(uint amIn,uint eth,bool direc,bytes memory _pools)internal view returns(uint hAmOut,uint poolCall){
        unchecked{
            uint hGasFee;
            for(uint p;p<_pools.length;p+=0x60){
                uint slot1;//=uint(bytes32(_pools[p:]));
                assembly{
                    slot1:=mload(add(add(_pools,p),0x40))
                }
                uint rIn;uint rOut;
                {
                    uint slot0;//=uint(bytes32(_pools[p:]));
                    assembly{
                        slot0:=mload(add(add(_pools,p),0x20))
                    }
                    rIn=slot0>>128;
                    rOut=uint128(slot0);
                }
                if(!direc){
                    (rIn,rOut)=(rOut,rIn);
                }
                {
                    uint slot2;//=uint(bytes32(_pools[p:]));
                    assembly{
                        slot2:=mload(add(add(_pools,p),0x60))
                    }
                    if(slot2!=0 && amIn+rIn>=(direc?(slot2>>128):uint128(slot2))){
                        continue;
                    }
                }
                uint amInXFee= amIn * uint24(slot1>>160);
                uint amOut = (amInXFee * rOut) / (rIn * 1e6 + amInXFee);
                uint gasFee = callGas(slot1) * tx.gasprice;
                if(eth!=0){
                    gasFee=(amOut*gasFee)/eth;
                }
                if(int(amOut-gasFee)<=int(hAmOut-hGasFee)){
                    continue;
                }
                if(FRP){
                    uint amOutX2 = amInXFee<<1;
                    amOutX2 = (amOutX2 * rOut) / (rIn * 1e6 + amOutX2);
                    if(int(amOutX2-gasFee)>int((amOut-gasFee)<<1)){
                        continue;
                    }
                }
                hGasFee=gasFee;
                hAmOut=amOut;
                poolCall=slot1;
            }
        }
    }

    function callGas(uint poolCall)internal pure returns(uint){
        return poolCall&PID_MASK==ALGB_PID?300000:100000;
    }

    function compress56bit(uint uncompressed)internal pure returns(uint){
        unchecked{
            uint temp=uncompressed;
            uint rsh;
            while(uint48(temp)!=temp){
                rsh+=8;
                temp>>=8;
            }
            temp<<=8;
            temp|=rsh;
            return temp;
        }
    }

    function poolInCalls(bytes memory calls,uint160 pool)internal pure returns(bool){
        // uint fees=uint24(slot1>>160);
        // uint160 pool=uint160(slot1);
        for (uint i=0x20; i <= calls.length; i += 0x20) {
            uint _poolCall;
            assembly{_poolCall:= mload(add(calls, i))}
            if (pool == uint160(_poolCall)){
                return true;
                // uint amounts[t0]=decompress56bit(_poolCall>>160);
                // uint amounts[t0]XFee=amounts[t0]*fees;
                // bool v2fee = uint8(slot1>>216)==1;
                // if(_poolCall&0x8000000000000000000000000000000000000000000000000000000000000000==0){
                //     r0-=(amounts[t0]XFee*r0)/(r1*1e6+amounts[t0]XFee);
                //     r1+=v2fee?amounts[t0]:(amounts[t0]XFee/1e6);
                // }else{
                //     r1-=(amounts[t0]XFee*r1)/(r0*1e6+amounts[t0]XFee);
                //     r0+=v2fee?amounts[t0]:(amounts[t0]XFee/1e6);
                // }
            }
        }
        return false;//(r0,r1);
    }
}
