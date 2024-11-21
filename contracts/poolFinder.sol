import "./router.sol";
import "@uniswap/v3-core/contracts/interfaces/IUniswapV3Pool.sol";
import "@uniswap/v3-core/contracts/interfaces/IUniswapV3Factory.sol";
import "@uniswap/v2-core/contracts/interfaces/IUniswapV2Pair.sol";
import "@uniswap/v2-core/contracts/interfaces/IUniswapV2Factory.sol";
import "@cryptoalgebra/integral-core/contracts/interfaces/IAlgebraPool.sol";
import "@cryptoalgebra/integral-core/contracts/interfaces/IAlgebraFactory.sol";

contract CPoolFinder{

    struct Protocol{
        address factory;
        uint8 id;
    }
    
    uint internal constant STATE_MASK = 0x7fffffff00000000000000000000000000000000000000000000000000000000;

    function findPools(uint minEth,address[] calldata tokens,Protocol[] calldata protocols)public view returns(bytes[][] memory pools){
        unchecked {
            pools=new bytes[][](tokens.length);
            for(uint t0;t0<tokens.length;t0++){
                for (uint t1; t1<tokens.length; t1++){
                    if(t0==t1 || tokens[t0]>tokens[t1]){
                        continue;
                    }
                    bytes memory _pools=findPoolsSingle(tokens[t0], tokens[t1], protocols);
                    if(_pools.length>0){
                        if(pools[t0].length==0){
                            pools[t0]=new bytes[](tokens.length);
                        }
                        pools[t0][t1]=_pools;
                    }
                }
            }
            (uint[] memory amounts,)=CRouter.findRoutesInt(2,0,minEth,pools);
            filterPools(amounts,pools);
        }
    }

    function findPoolsSingle(address token0,address token1,Protocol[] calldata protocols)public view returns(bytes memory pools){
        unchecked{
            if(token0>token1){
                (token0,token1)=(token1,token0);
            }
            assembly{
                pools:=mload(0x40)
                mstore(0x40,add(pools,0x20))
            }
            for(uint i; i<protocols.length;i++){
                if(protocols[i].id==0){
                    mstoreUniV3Pool(protocols[i].factory,token0,token1,100);
                    mstoreUniV3Pool(protocols[i].factory,token0,token1,500);
                    mstoreUniV3Pool(protocols[i].factory,token0,token1,2500);
                    mstoreUniV3Pool(protocols[i].factory,token0,token1,3000);
                    mstoreUniV3Pool(protocols[i].factory,token0,token1,10000);
                }else if(protocols[i].id==1){
                    mstoreUniV2Pool(protocols[i].factory,token0,token1);
                }else if(protocols[i].id==2){
                    mstoreAlgbPool(protocols[i].factory,token0,token1);
                }
            }
            uint len;
            assembly{len:=sub(mload(0x40),add(pools,0x20))}
            if(len==0){
                delete pools;
            }else{
                assembly{mstore(pools,len)}
            }
        }
    }

    function filterPools(uint[] memory fAmounts, bytes[][] memory pools) internal pure{
        unchecked{
            for (uint t0; t0 < pools.length; t0++){
                for (uint t1; t1 < pools[t0].length; t1++){
                    if(t0==t1){
                        continue;
                    }
                    bytes memory _pools=pools[t0][t1];
                    if(_pools.length==0){
                        continue;
                    }
                    uint _len;
                    uint p;
                    while(p<_pools.length){
                        uint slot0;uint slot1;
                        assembly{
                            p:=add(p,0x20)
                            slot0:=mload(add(_pools,p))
                            p:=add(p,0x20)
                            slot1:=mload(add(_pools,p))
                        }
                        uint rt0=slot0>>128;
                        uint rt1=uint128(slot0);
                        uint fee=1e6-uint24(slot1>>160);
                        uint amt0=fAmounts[t1]*fee;
                        amt0=(amt0 * rt0) / (rt1 * 1e6 + amt0);
                        uint amt1=fAmounts[t0]*fee;
                        amt1=(amt1 * rt1) / (rt0 * 1e6 + amt1);
                        if((amt1+(amt1>>1)<fAmounts[t1]) && (amt0+(amt0>>1)<fAmounts[t0])){
                            continue;
                        }
                        assembly{
                            _len:=add(_len,0x20)
                            mstore(add(_pools,_len),slot0)
                            _len:=add(_len,0x20)
                            mstore(add(_pools,_len),slot1)
                        }
                    }
                    if(_len>0){
                        assembly{
                            mstore(_pools,_len)
                        }
                    }else{
                        delete pools[t0][t1];
                    }
                }
            }
            for(uint t0;t0<pools.length;t0++){
                bool b;
                for(uint t1;t1<pools[t0].length;t1++){
                    if(pools[t0][t1].length>0){
                        b=true;
                    }
                }
                if(!b){
                    delete pools[t0];
                }
            }
        }
    }

    function mstoreUniV2Pool(address factory,address t0,address t1) internal view{
        unchecked{
            bytes32 fmp;
            assembly{fmp:=mload(0x40)}
            address pool=IUniswapV2Factory(factory).getPair(t0,t1);
            if(pool.code.length!=0){
                uint reserve0; uint reserve1;bytes32 stateHash;
                bytes4 sel=IUniswapV2Pair(pool).getReserves.selector;
                assembly{
                    mstore(fmp,sel)
                    pop(staticcall(gas(), pool, fmp, 0x04, fmp, 0x40))
                    reserve0:=mload(fmp)
                    reserve1:=mload(add(fmp,0x20))
                }
                if(reserve0>0&&reserve1>0){
                    uint8 id=1;
                    assembly{
                        stateHash:=keccak256(fmp,0x40)
                        mstore(fmp,or(shl(128,reserve0),reserve1))
                        fmp:=add(fmp,0x20)
                        mstore(fmp,or(and(stateHash,STATE_MASK),or(shl(216,id),or(shl(160,3000),pool))))
                        fmp:=add(fmp,0x20)
                    }
                }
            }
            assembly{mstore(0x40,fmp)}
        }
    }

    function mstoreUniV3Pool(address factory,address t0,address t1,uint24 fee)internal view{
        unchecked{
            bytes32 fmp;
            assembly{fmp:=mload(0x40)}
            address pool=IUniswapV3Factory(factory).getPool(t0,t1,fee);
            if(pool.code.length!=0){
                uint liquidity=IUniswapV3Pool(pool).liquidity();
                if(liquidity>0){
                    uint sqrtPX64;
                    bytes4 sel=IUniswapV3Pool(pool).slot0.selector;
                    assembly{
                        mstore(fmp,sel)
                        pop(staticcall(gas(), pool, fmp, 0x04, fmp, 0x40))
                        sqrtPX64 := shr(32,mload(fmp))
                    }
                    uint reserve0=(liquidity<<64)/(sqrtPX64+1);
                    uint reserve1=(liquidity*sqrtPX64)>>64;
                    if(reserve0>0&&reserve1>0){
                        uint8 id=0;
                        assembly{
                            let t:=mload(add(fmp,0x20))
                            let stateHash:=keccak256(fmp,0x20)
                            mstore(fmp,or(shl(128,reserve0),reserve1))
                            fmp:=add(fmp,0x20)
                            mstore(fmp,or(and(stateHash,STATE_MASK),or(shl(216,0),or(shl(176,and(t,0xffffff)),or(shl(160,fee),pool)))))
                            fmp:=add(fmp,0x20)
                        }
                    }
                }
            }
            assembly{mstore(0x40,fmp)}
        }
    }

    function mstoreAlgbPool(address factory,address t0,address t1)internal view{
        unchecked{
            bytes32 fmp;
            assembly{fmp:=mload(0x40)}
            address pool = IAlgebraFactory(factory).poolByPair(t0,t1);
            if(pool.code.length!=0){
                uint liquidity =IAlgebraPool(pool).liquidity();
                if(liquidity>0){
                    uint sqrtPX64;
                    bytes4 sel=IAlgebraPool(pool).globalState.selector;
                    assembly{
                        mstore(fmp,sel)
                        pop(staticcall(gas(), pool, fmp, 0x04, fmp, 0x60))
                        sqrtPX64 := shr(32,mload(fmp))
                    }
                    uint reserve0=(liquidity<<64)/(sqrtPX64+1);
                    uint reserve1=(liquidity*sqrtPX64)>>64;
                    if(reserve0>0&&reserve1>0){
                        uint8 id=2;
                        assembly{
                            let t:=mload(add(fmp,0x20))
                            let fee:=mload(add(fmp,0x40))
                            let stateHash:=keccak256(fmp,0x20)
                            mstore(fmp,or(shl(128,reserve0),reserve1))
                            fmp:=add(fmp,0x20)
                            mstore(fmp,or(and(stateHash,STATE_MASK),or(shl(216,id),or(shl(176,and(t,0xffffff)),or(shl(160,fee),pool)))))
                            fmp:=add(fmp,0x20)
                        }
                    }
                }
            }
            assembly{mstore(0x40,fmp)}
        }
    }
}


