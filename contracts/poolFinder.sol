contract PoolFinder{

    struct Protocol{
        bytes32 initCode;
        address factory;
    }

    struct Protocols{
        Protocol[] uniV2;
        Protocol[] uniV3;
        Protocol[] algebraV3;
    }

    int internal constant MIN_TICK = -887272;
    int internal constant MAX_TICK = 887272;
    bytes32 internal constant B4_MASK = 0xffffffff00000000000000000000000000000000000000000000000000000000;
    bytes32 internal constant ADDR_MASK = 0x000000000000000000000000ffffffffffffffffffffffffffffffffffffffff;

    function findPools(address token0,address token1,Protocols calldata protocols)public returns(bytes memory pools){
        unchecked {
            if(token0>token1)
                (token0,token1)=(token1,token0);
            assembly{
                pools:=mload(0x40)
                mstore(0x40,add(pools,0x20))
            }
            for(uint i; i<protocols.uniV2.length;i++){
                mstoreUniV2Pool(protocols.uniV2[i],token0,token1);
            }
            for(uint i; i<protocols.uniV3.length;i++){
                mstoreUniV3Pool(protocols.uniV3[i],token0,token1,100,1);
                mstoreUniV3Pool(protocols.uniV3[i],token0,token1,500,10);
                mstoreUniV3Pool(protocols.uniV3[i],token0,token1,3000,60);
                mstoreUniV3Pool(protocols.uniV3[i],token0,token1,10000,200);
            }
            for(uint i; i<protocols.algebraV3.length;i++){
                mstoreAlgebraV3Pool(protocols.algebraV3[i],token0,token1);
            }
            uint len;
            assembly{
                len:=sub(sub(mload(0x40),pools),0x20)
            }
            if(len>0){
                assembly{
                    mstore(pools,len)
                }
            }
        }
    }

    function mstoreUniV2Pool(Protocol calldata protocol,address t0,address t1) internal{
        unchecked{
            bytes32 fmp;
            assembly{fmp:=mload(0x40)}
            address pool=address(uint160(uint(keccak256(abi.encodePacked(hex'ff',protocol.factory, keccak256(abi.encodePacked(t0, t1)) ,protocol.initCode)))));
            if(pool.code.length>0){
                uint reserve0; uint reserve1;bytes32 stateHash;
                assembly{
                    mstore(fmp,0x0902f1ac00000000000000000000000000000000000000000000000000000000)
                    pop(call(gas(), pool, 0, fmp, 0x04, fmp, 0x40))
                    reserve0:=mload(fmp)
                    reserve1:=mload(add(fmp,0x20))
                }
                if(reserve0>0 && reserve1>0){
                    assembly{
                        stateHash:=keccak256(fmp,0x20)
                        mstore(fmp,or(shl(128,reserve0),reserve1))
                        fmp:=add(fmp,0x20)
                        mstore(fmp,or(and(stateHash,B4_MASK),or(shl(216,1),or(shl(160,997000),and(pool,ADDR_MASK)))))
                        fmp:=add(fmp,0x40)
                    }
                }
            }
            assembly{mstore(0x40,fmp)}
        }
    }

    function mstoreUniV3Pool(Protocol calldata protocol,address t0,address t1,uint fee,int s)internal {
        unchecked{
            bytes32 fmp;
            assembly{fmp:=mload(0x40)}
            address pool=address(uint160(uint(keccak256(abi.encodePacked(hex'ff',protocol.factory, keccak256(abi.encode(t0, t1,fee)),protocol.initCode)))));
            if(pool.code.length>0){
                uint liquidity=IUniV3Pool(pool).liquidity();
                if(liquidity>2){
                    int t;uint sqrtPX64;bytes32 stateHash;
                    assembly{
                        mstore(fmp,0x3850c7bd00000000000000000000000000000000000000000000000000000000)
                        pop(call(gas(), pool, 0, fmp, 0x04, fmp, 0x40))
                        sqrtPX64 := shr(32,mload(fmp))
                        t:=mload(add(fmp,0x20))
                        stateHash:=keccak256(fmp,0x20)
                    }
                    (uint reserve0,uint reserve1,uint reserve0Limit,uint reserve1Limit)=reserves(liquidity,sqrtPX64,t,s);
                    assembly{
                        mstore(fmp,or(shl(128,reserve0),reserve1))
                        fmp:=add(fmp,0x20)
                        mstore(fmp,or(and(stateHash,B4_MASK),or(shl(160,sub(1000000,fee)),and(pool,ADDR_MASK))))
                        fmp:=add(fmp,0x20)
                        mstore(fmp,or(shl(128,reserve0Limit),reserve1Limit))
                        fmp:=add(fmp,0x20)
                    }
                }
            }
            assembly{mstore(0x40,fmp)}
        }
    }

    function mstoreAlgebraV3Pool(Protocol calldata protocol,address t0,address t1)internal {
        unchecked{
            bytes32 fmp;
            assembly{fmp:=mload(0x40)}
            address pool =  address(uint160(uint(keccak256(abi.encodePacked(hex'ff',protocol.factory,keccak256(abi.encode(t0, t1)),protocol.initCode)))));
            if(pool.code.length>0){
                uint liquidity =IAlgebraV3Pool(pool).liquidity();
                if(liquidity>2){
                    int t;uint sqrtPX64;bytes32 stateHash;
                    assembly{
                        mstore(fmp,0xe76c01e400000000000000000000000000000000000000000000000000000000)
                        pop(call(gas(), pool, 0, fmp, 0x04, fmp, 0x60))
                        sqrtPX64 := shr(32,mload(fmp))
                        t:=mload(add(fmp,0x20))
                        stateHash:=keccak256(fmp,0x20)
                    }
                    (uint reserve0,uint reserve1,uint reserve0Limit,uint reserve1Limit)=reserves(liquidity,sqrtPX64,t,60);
                    assembly{
                        let fee:=mload(add(fmp,0x40))
                        mstore(fmp,or(shl(128,reserve0),reserve1))
                        fmp:=add(fmp,0x20)
                        mstore(fmp,or(and(stateHash,B4_MASK),or(shl(216,2),or(shl(160,sub(1000000,fee)),and(pool,ADDR_MASK)))))
                        fmp:=add(fmp,0x20)
                        mstore(fmp,or(shl(128,reserve0Limit),reserve1Limit))
                        fmp:=add(fmp,0x20)
                    }
                }
            }
            assembly{mstore(0x40,fmp)}
        }
    }


    function reserves(uint liquidity,uint sqrtPX64,int t, int s)internal pure returns (uint reserve0,uint reserve1, uint reserve0Limit,uint reserve1Limit){
        unchecked{
            reserve0=(liquidity<<64)/sqrtPX64;
            reserve1=(liquidity*sqrtPX64)>>64;
            int tl;
            assembly {tl := mul(sub(sdiv(t, s), and(slt(t, 0), smod(t, s))), s)}
            int tu=tl+s;
            if(tl<MIN_TICK){
                tl=MIN_TICK;
            }
            else if(tu>MAX_TICK){
                tu=MAX_TICK;
            }
            reserve0Limit=(liquidity<<64)/(tSqrtPX64(tl)+1);
            reserve1Limit=((liquidity*tSqrtPX64(tu))>>64);
            if(reserve0Limit<reserve0)
                reserve0Limit=reserve0;
            if(reserve1Limit<reserve1)
                reserve1Limit=reserve1;
        }
    }

    function tSqrtPX64(int tick) internal pure returns (uint sqrtPriceX64) {
        unchecked {
            uint256 absTick;
            assembly {
                let mask := sar(255, tick)
                absTick := xor(mask, add(mask, tick))
            }
            uint256 price;
            assembly {
                price := xor(shl(128, 1), mul(xor(shl(128, 1), 0xfffcb933bd6fad37aa2d162d1a594001), and(absTick, 0x1)))
            }
            if (absTick & 0x2 != 0) price = (price * 0xfff97272373d413259a46990580e213a) >> 128;
            if (absTick & 0x4 != 0) price = (price * 0xfff2e50f5f656932ef12357cf3c7fdcc) >> 128;
            if (absTick & 0x8 != 0) price = (price * 0xffe5caca7e10e4e61c3624eaa0941cd0) >> 128;
            if (absTick & 0x10 != 0) price = (price * 0xffcb9843d60f6159c9db58835c926644) >> 128;
            if (absTick & 0x20 != 0) price = (price * 0xff973b41fa98c081472e6896dfb254c0) >> 128;
            if (absTick & 0x40 != 0) price = (price * 0xff2ea16466c96a3843ec78b326b52861) >> 128;
            if (absTick & 0x80 != 0) price = (price * 0xfe5dee046a99a2a811c461f1969c3053) >> 128;
            if (absTick & 0x100 != 0) price = (price * 0xfcbe86c7900a88aedcffc83b479aa3a4) >> 128;
            if (absTick & 0x200 != 0) price = (price * 0xf987a7253ac413176f2b074cf7815e54) >> 128;
            if (absTick & 0x400 != 0) price = (price * 0xf3392b0822b70005940c7a398e4b70f3) >> 128;
            if (absTick & 0x800 != 0) price = (price * 0xe7159475a2c29b7443b29c7fa6e889d9) >> 128;
            if (absTick & 0x1000 != 0) price = (price * 0xd097f3bdfd2022b8845ad8f792aa5825) >> 128;
            if (absTick & 0x2000 != 0) price = (price * 0xa9f746462d870fdf8a65dc1f90e061e5) >> 128;
            if (absTick & 0x4000 != 0) price = (price * 0x70d869a156d2a1b890bb3df62baf32f7) >> 128;
            if (absTick & 0x8000 != 0) price = (price * 0x31be135f97d08fd981231505542fcfa6) >> 128;
            if (absTick & 0x10000 != 0) price = (price * 0x9aa508b5b7a84e1c677de54f3e99bc9) >> 128;
            if (absTick & 0x20000 != 0) price = (price * 0x5d6af8dedb81196699c329225ee604) >> 128;
            if (absTick & 0x40000 != 0) price = (price * 0x2216e584f5fa1ea926041bedfe98) >> 128;
            if (absTick & 0x80000 != 0) price = (price * 0x48a170391f7dc42444e8fa2) >> 128;
            assembly {
                if sgt(tick, 0) { price := div(not(0), price) }
                sqrtPriceX64 := shr(64, price)
            }
        }
    }

}

//interfaces

interface IUniV2Pool{
    function getReserves()external view returns(uint reserve0, uint reserve1, uint blockTimestampLast);
}

interface IUniV3Pool{
    function slot0() external view returns(uint sqrtPX96, int t, uint observationIndex, uint observationCardinality, uint observationCardinalityNext, uint feeProtocol, bool unlocked);
    function liquidity() external view returns(uint liquidity);
}

interface IAlgebraV3Pool{
    function globalState() external view returns(uint sqrtPX96, int t, uint fee, uint timePointdex, uint comunityFeet0, uint comunityFeeT1, bool unlocked);
    function tickSpacing() external view returns(int s);
    function liquidity() external view returns(uint liquidity);
}


